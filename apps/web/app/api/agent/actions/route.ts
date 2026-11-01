import {proxyAgent} from "../../../../lib/agent-proxy";

export async function GET(request: Request) {
  return proxyAgent(request, "actions", "GET");
}
