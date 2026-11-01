// Local deterministic model fixture for Docker integration tests only.
// No actual AI is used, no external API requests are sent.
"use strict";
const http = require("node:http");
const host = "0.0.0.0";
const port = 8999;
const respond = (res, status, body) => {
  res.writeHead(status, {"Content-Type": "application/json", "Cache-Control": "no-store"});
  res.end(JSON.stringify(body));
};
http.createServer((req, res) => {
  if (req.method === "GET" && req.url === "/health") {
    return respond(res, 200, {status: "ok"});
  }
  if (req.method !== "POST" || req.url !== "/v1/chat/completions") {
    return respond(res, 404, {error: {message: "Not found"}});
  }
  let data = "";
  req.on("data", chunk => {
    data += chunk;
    if (data.length > 150000) req.destroy();
  });
  req.on("end", () => {
    let payload;
    try { payload = JSON.parse(data); } catch { return respond(res, 400, {error: {message: "Bad JSON"}}); }
    const messages = payload.messages || [];
    const toolReplied = messages.some(message => message.role === "tool");
    if (toolReplied) {
      return respond(res, 200, {choices: [{message: {role: "assistant", content: "The requested tool was checked. Review any proposed action before confirming."}}]});
    }
    const prompt = [...messages].reverse().find(message => message.role === "user")?.content || "";
    const prepare = prompt.toLowerCase().includes("prepare");
    const dateMatches = prompt.match(/\d{4}-\d{2}-\d{2}/g) || [];
    const tool = prepare ? "submit_leave_request" : "get_leave_balance";
    const args = prepare ? {
      leave_type: "annual", start_date: dateMatches[0], end_date: dateMatches[1],
      partial_day: "full", reason: "CI disposable test request"
    } : {};
    return respond(res, 200, {choices: [{message: {
      role: "assistant",
      tool_calls: [{id:"mock-call-001",type:"function",function:{name:tool,arguments:JSON.stringify(args)}}]
    }}]});
  });
}).listen(port, host);
