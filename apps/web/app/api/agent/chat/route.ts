import {proxyAgent} from "../../../../lib/agent-proxy";

export async function POST(request: Request) {
  return proxyAgent(request, "chat");
}
