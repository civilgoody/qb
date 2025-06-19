import { NextResponse } from "next/server";

export async function GET(req: Request) {
  return NextResponse.json({ msg: "Q is live" }, { status: 200 });
}
