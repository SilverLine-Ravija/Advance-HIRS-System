import {cookies} from "next/headers";
import {NextResponse} from "next/server";
import {apiBaseURL, SESSION_COOKIE} from "../../../../lib/api";

/** Sign-out revokes the bearer token in Go before clearing the HttpOnly cookie. */
export async function POST(request: Request) {
  const origin = request.headers.get("origin");
  if (request.headers.get("sec-fetch-site") === "cross-site") {
    return NextResponse.json({error: {code: "forbidden", message: "Cross-site sign out denied"}}, {status: 403});
  }
  if (origin) {
    try {
      if (new URL(origin).host !== new URL(request.url).host) {
        return NextResponse.json({error: {code: "forbidden", message: "Cross-origin sign out denied"}}, {status: 403});
      }
    } catch {
      return NextResponse.json({error: {code: "forbidden", message: "Invalid origin"}}, {status: 403});
    }
  }
  const cookieStore = await cookies();
  const token = cookieStore.get(SESSION_COOKIE)?.value;
  if (!token) {
    return NextResponse.json({authenticated: false}, {headers: {"Cache-Control": "no-store"}});
  }
  let status: number;
  try {
    const upstream = await fetch(`${apiBaseURL()}/api/v1/auth/logout`, {
      method: "POST",
      headers: {Authorization: `Bearer ${token}`},
      cache: "no-store",
      signal: AbortSignal.timeout(10_000),
    });
    status = upstream.status;
  } catch {
    return NextResponse.json({error: {code: "revoke_unavailable", message: "Could not verify secure sign out. Please retry."}}, {status: 503});
  }
  // An already invalid/expired token is safe to forget locally.
  if (status !== 200 && status !== 401) {
    return NextResponse.json({error: {code: "revoke_unavailable", message: "Could not verify secure sign out. Please retry."}}, {status: 503});
  }
  cookieStore.delete(SESSION_COOKIE);
  return NextResponse.json({authenticated: false}, {headers: {"Cache-Control": "no-store"}});
}
