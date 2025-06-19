import { NextResponse } from "next/server";
import { handleUpload } from "./uploads/zip";

// API Route handler
export async function POST(req: Request) {
  try {
    const res = await handleUpload(await req.formData());
    return NextResponse.json({ res });
  } catch (error: any) {
    console.error("Error uploading file:", error);
    if (error.response) {
      console.error("Response data:", error.response.data);
    }
    return NextResponse.json({
      status: 500,
      error: "Failed to upload and send file to Telegram.",
    });
  }
}
