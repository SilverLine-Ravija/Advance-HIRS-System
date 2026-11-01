import {cookies} from "next/headers";
import {NextResponse} from "next/server";
import {SESSION_COOKIE} from "./api";

type AgentEndpoint = "chat" | "confirm" | "actions" | `actions/${string}/dismiss`;

/** Server-only agent proxy. Never exposes HRIS JWT or private agent URL to browser JS. */
export async function proxyAgent(request: Request, endpoint: AgentEndpoint, method: "GET" | "POST" = "POST") {
  const cookieStore = await cookies();
  const token = cookieStore.get(SESSION_COOKIE)?.value;
  if (!token) {
    return NextResponse.json({error: {code: "unauthorized", message: "Sign in first"}}, {status: 401});
  }
  if (method === "POST" && request.headers.get("content-type")?.split(";")[0] !== "application/json") {
    return NextResponse.json({error: {message: "Expected JSON"}}, {status: 415});
  }
  const origin = request.headers.get("origin");
  if (method === "POST" && origin) {
    try {
      if (new URL(origin).host !== new URL(request.url).host) {
        return NextResponse.json({error: {message: "Cross-origin mutation denied"}}, {status: 403});
      }
    } catch {
      return NextResponse.json({error: {message: "Invalid request origin"}}, {status: 403});
    }
  }
  const body = method === "POST" ? await request.text() : undefined;
  if (body && body.length > 64_000) {
    return NextResponse.json({error: {message: "Request too large"}}, {status: 413});
  }
  const query = method === "GET" ? new URL(request.url).search : "";
  const target = `${process.env.AGENT_INTERNAL_URL || "http://127.0.0.1:8090"}/api/agent/${endpoint}${query}`;
  try {
    const response = await fetch(target, {
      method,
      headers: {Authorization: `Bearer ${token}`, "Content-Type": "application/json"},
      body,
      cache: "no-store",
      signal: AbortSignal.timeout(60_000),
    });
    const type = response.headers.get("content-type") || "application/json";
    return new NextResponse(await response.arrayBuffer(), {
      status: response.status,
      headers: {"Content-Type": type, "Cache-Control": "no-store"},
    });
  } catch {
    return NextResponse.json({error: {message: "HR assistant is unavailable"}}, {status: 502});
  }
}
