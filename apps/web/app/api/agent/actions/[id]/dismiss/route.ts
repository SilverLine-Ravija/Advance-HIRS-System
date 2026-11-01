import {NextResponse} from "next/server";
import {proxyAgent} from "../../../../../../lib/agent-proxy";

export async function POST(request: Request, context: {params: Promise<{id: string}>}) {
  const {id} = await context.params;
  if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(id)) {
    return NextResponse.json({error: {code: "invalid_action", message: "Invalid action ID"}}, {status: 400});
  }
  return proxyAgent(request, `actions/${id}/dismiss`, "POST");
}
