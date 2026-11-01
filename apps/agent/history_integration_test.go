package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

func TestActionHistoryDismissAndActorIsolation(t *testing.T) {
    dsn := os.Getenv("AGENT_TEST_DATABASE_URL")
    if dsn == "" { t.Skip("AGENT_TEST_DATABASE_URL is required for integration testing") }
    ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil { t.Fatal(err) }
    defer pool.Close()

    slug := fmt.Sprintf("history-test-%d", time.Now().UnixNano())
    var orgID, actorID, otherID string
    if err := pool.QueryRow(ctx,`INSERT INTO organizations(name,slug) VALUES ('History Test',$1) RETURNING id::text`,slug).Scan(&orgID); err != nil { t.Fatal(err) }
    defer pool.Exec(context.Background(), `DELETE FROM organizations WHERE id=$1::uuid`, orgID)
    for _, entry := range []struct {email string; into *string}{
        {slug+"-actor@example.invalid",&actorID},
        {slug+"-other@example.invalid",&otherID},
    } {
        if err := pool.QueryRow(ctx, `INSERT INTO users(email,password_hash,display_name)
            VALUES ($1,'integration-test-hash','History Test') RETURNING id::text`,entry.email).Scan(entry.into); err != nil { t.Fatal(err) }
        defer pool.Exec(context.Background(), `DELETE FROM users WHERE id=$1::uuid`, *entry.into)
    }
    current := principal{OrganizationID:orgID,UserID:actorID,Role:"employee"}
    other := principal{OrganizationID:orgID,UserID:otherID,Role:"employee"}
    upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
        if r.URL.Path != "/api/v1/auth/me" { http.Error(w,"unexpected operation",500); return }
        caller := current
        if r.Header.Get("Authorization") == "Bearer history-other-token-long" { caller = other }
        _ = json.NewEncoder(w).Encode(map[string]any{"user":caller})
    }))
    defer upstream.Close()
    s := &server{config:config{apiURL:upstream.URL},client:&http.Client{Timeout:2*time.Second},pool:pool,limits:map[string]requestCount{}}
    args := map[string]any{"leave_type":"annual","start_date":"2026-11-16","end_date":"2026-11-17","reason":"Integration test"}
    plan, err := planTool(current,"submit_leave_request",args)
    if err != nil { t.Fatal(err) }
    preview, err := s.stageAction(ctx,current,plan,args)
    if err != nil { t.Fatal(err) }

    request := func(method,path,token string,body []byte) *httptest.ResponseRecorder {
        req := httptest.NewRequest(method,path,bytes.NewReader(body))
        req.Header.Set("Authorization","Bearer "+token)
        recorder := httptest.NewRecorder()
        switch {
        case method==http.MethodGet: s.actionsHistory(recorder,req)
        case method==http.MethodPost:
            req.SetPathValue("id",preview.ID)
            s.dismissAction(recorder,req)
        }
        return recorder
    }
    ownHistory := request("GET","/api/agent/actions?limit=10","history-actor-token-long",nil)
    if ownHistory.Code != 200 { t.Fatalf("own history: %d %s",ownHistory.Code,ownHistory.Body.String()) }
    var list struct {Data []actionHistoryItem `json:"data"`}
    if err := json.Unmarshal(ownHistory.Body.Bytes(),&list); err != nil { t.Fatal(err) }
    if len(list.Data) != 1 || list.Data[0].ID != preview.ID || list.Data[0].Status != "pending" { t.Fatalf("unexpected own history: %+v",list.Data) }
    otherHistory := request("GET","/api/agent/actions","history-other-token-long",nil)
    var otherList struct {Data []actionHistoryItem `json:"data"`}
    if err := json.Unmarshal(otherHistory.Body.Bytes(),&otherList);err != nil { t.Fatal(err) }
    if len(otherList.Data) != 0 { t.Fatalf("other actor can read action history: %+v",otherList.Data) }
    forbidden := request("POST","/api/agent/actions/"+preview.ID+"/dismiss","history-other-token-long",[]byte("{}"))
    if forbidden.Code != 409 { t.Fatalf("other actor could dismiss proposal: %d",forbidden.Code) }
    dismissed := request("POST","/api/agent/actions/"+preview.ID+"/dismiss","history-actor-token-long",[]byte("{}"))
    if dismissed.Code != 200 { t.Fatalf("dismiss failed: %d %s",dismissed.Code,dismissed.Body.String()) }
    history := request("GET","/api/agent/actions","history-actor-token-long",nil)
    if err := json.Unmarshal(history.Body.Bytes(),&list); err != nil { t.Fatal(err) }
    if len(list.Data)!=1 || list.Data[0].Status!="dismissed" {t.Fatalf("wrong dismiss state: %+v",list.Data)}
    again := request("POST","/api/agent/actions/"+preview.ID+"/dismiss","history-actor-token-long",[]byte("{}"))
    if again.Code != 409 {t.Fatalf("dismiss should be single-use: %d",again.Code)}
    confirmReq := httptest.NewRequest("POST","/api/agent/confirm",bytes.NewBufferString(fmt.Sprintf(`{"action_id":%q}`,preview.ID)))
    confirmReq.Header.Set("Authorization","Bearer history-actor-token-long")
    confirmed := httptest.NewRecorder()
    s.confirm(confirmed,confirmReq)
    if confirmed.Code != 409 {t.Fatalf("dismissed proposal was confirmable: %d",confirmed.Code)}

    invalid := request("GET","/api/agent/actions?limit=1000","history-actor-token-long",nil)
    if invalid.Code != 400 {t.Fatalf("invalid history limit accepted: %d",invalid.Code)}
}
