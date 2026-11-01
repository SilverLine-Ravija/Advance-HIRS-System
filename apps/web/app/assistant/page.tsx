"use client";

import {FormEvent, useEffect, useState} from "react";
import Link from "next/link";
import styles from "./assistant.module.css";

type Turn = {role: "user" | "assistant"; content: string};
type Proposal = {id: string; tool: string; preview: string; expires_at: string};
type Response = {reply: string; pending_actions: Proposal[]};

const examples = [
  "How much annual leave do I have left?",
  "What are our company holidays?",
  "Show my attendance and any corrections awaiting review.",
  "Summarize the pending leave requests I'm allowed to review.",
];

async function post<T>(endpoint: string, body: unknown): Promise<T> {
  const response = await fetch(endpoint, {
    method: "POST", headers: {"Content-Type": "application/json"},
    cache: "no-store", body: JSON.stringify(body),
  });
  const json = await response.json();
  if (!response.ok) throw new Error(json?.error?.message || json?.message || `Request failed (${response.status})`);
  return json as T;
}

export default function AssistantPage() {
  const [turns, setTurns] = useState<Turn[]>([]);
  const [proposals, setProposals] = useState<Proposal[]>([]);
  const [draft, setDraft] = useState("");
  const [busy, setBusy] = useState(false);
  const [confirming, setConfirming] = useState("");
  const [error, setError] = useState("");
  const [signedIn, setSignedIn] = useState<boolean | null>(null);

  useEffect(() => {
    fetch("/api/hr/auth/me", {cache: "no-store"})
      .then(r => setSignedIn(r.ok))
      .catch(() => setSignedIn(false));
  }, []);

  async function send(event?: FormEvent, suggestion?: string) {
    event?.preventDefault();
    const message = (suggestion || draft).trim();
    if (!message || busy || confirming || proposals.length) return;
    const history = turns.slice(-8);
    setBusy(true);
    setError("");
    setDraft("");
    setTurns(old => [...old, {role: "user" as const, content: message}]);
    try {
      const result = await post<Response>("/api/agent/chat", {message, history});
      setTurns(old => [...old, {role: "assistant" as const, content: result.reply}]);
      setProposals(result.pending_actions || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to reach the assistant");
      setTurns(old => [...old, {role: "assistant" as const, content: "I couldn't finish that request. No changes were confirmed."}]);
    } finally {
      setBusy(false);
    }
  }

  async function confirm(action: Proposal) {
    if (busy || confirming) return;
    setConfirming(action.id);
    setError("");
    try {
      const response = await post<{status: string; message?: string}>("/api/agent/confirm", {action_id: action.id});
      setTurns(old => [...old, {role: "assistant" as const, content: `${response.message || "Action processed"} (status: ${response.status}). See your HRIS record for the final result.`}]);
      setProposals(old => old.filter(item => item.id !== action.id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not confirm action");
      // Failed and ambiguous executions are not safe to resubmit from this card.
      setProposals(old => old.filter(item => item.id !== action.id));
    } finally {
      setConfirming("");
    }
  }

  return <main className={styles.shell}>
    <header className={styles.header}>
      <Link href="/" className={styles.back}>← Dashboard</Link>
      <span className={styles.flag}>HRIS / AI workspace</span>
    </header>
    <section className={styles.intro}>
      <p className={styles.kicker}>Advance HRIS</p>
      <h1>Meet your HR assistant.</h1>
      <p>Ask about leave, attendance and team workflows. Review and explicitly confirm every proposed change.</p>
    </section>
    <section className={styles.workspace} aria-label="HR assistant chat">
      {signedIn === false && <div className={styles.notice}>You need to <Link href="/">sign in on the dashboard</Link> to use your HR records.</div>}
      {turns.length === 0 && <div className={styles.welcome}>
        <h2>What would you like to do?</h2>
        <p>Your assistant can use only the data and tools allowed by your HRIS role.</p>
        <div className={styles.examples}>{examples.map(value => <button key={value} type="button" onClick={() => void send(undefined, value)} disabled={busy || signedIn !== true}>{value}</button>)}</div>
      </div>}
      <div className={styles.turns} role="log" aria-live="polite">
        {turns.map((turn, idx) => <div className={`${styles.turn} ${turn.role === "user" ? styles.user : styles.bot}`} key={idx}>
          <strong>{turn.role === "user" ? "You" : "HR Assistant"}</strong>
          <p>{turn.content}</p>
        </div>)}
        {busy && <p className={styles.working}>Checking your authorized HR records…</p>}
      </div>
      {proposals.length > 0 && <div className={styles.proposals}>
        <h2>Confirmation required</h2>
        <p>These changes have not been executed. Check all details before you confirm.</p>
        {proposals.map(action => <article className={styles.proposal} key={action.id}>
          <span className={styles.tool}>{action.tool.replaceAll("_", " ")}</span>
          <p>{action.preview}</p>
          <small>Expires {new Date(action.expires_at).toLocaleString()}</small>
          <div className={styles.actions}>
            <button type="button" onClick={() => void confirm(action)} disabled={Boolean(confirming)}> {confirming === action.id ? "Submitting…" : "Confirm action"}</button>
            <button type="button" className={styles.secondary} onClick={() => setProposals(old => old.filter(item => item.id !== action.id))} disabled={Boolean(confirming)}>Dismiss</button>
          </div>
        </article>)}
      </div>}
      {error && <div role="alert" className={styles.error}>{error}</div>}
      <form className={styles.form} onSubmit={e => void send(e)}>
        <label htmlFor="assistant-message" className={styles.srOnly}>Message the HR assistant</label>
        <textarea id="assistant-message" rows={2} value={draft} maxLength={3000} onChange={e => setDraft(e.target.value)} placeholder={proposals.length ? "Confirm or dismiss pending changes first" : "Ask about your HR workflows…"} disabled={busy || signedIn !== true || proposals.length > 0}/>
        <button type="submit" disabled={!draft.trim() || busy || signedIn !== true || proposals.length > 0}>Send</button>
      </form>
      <p className={styles.disclaimer}>AI can make mistakes. Verify sensitive HR information before taking action. Your HRIS permissions still apply.</p>
    </section>
  </main>;
}
