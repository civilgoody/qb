import TelegramBot from "node-telegram-bot-api";
import { z } from "zod";
import AdmZip from "adm-zip";
import { nanoid } from "nanoid";

// Initialize Telegram Bot with your API Token
const bot = new TelegramBot(process.env.TELEGRAM_TOKEN!, {
  polling: false,
  request: {
    url: "https://api.telegram.org",
    agentOptions: {
      keepAlive: true,
      family: 4,
    },
  },
});
const chatId = "-1002413392883";
const MAX_FILE_SIZE = 50 * 1024 * 1024;
const urlSchema = z
  .string()
  .url()
  .refine((link) => link !== "https://drive.google.com", {
    message: "Invalid link",
  });

export async function handleUpload(formData: FormData) {
  try {
    const name = formData.get("name") as string;
    const files = formData.getAll("files") as File[];
    const link = formData.get("link") as string;

    // Ensure name and link are valid strings
    if (!name || typeof name !== "string") {
      return { status: 400, msg: "Invalid name" };
    }

    let caption = buildCaption(name, link);

    // Handle file upload if files are provided
    if (files.length > 0) {
      try {
        return await handleFileUpload(files, caption);
      } catch (fileError) {
        return { status: 500, msg: "File upload failed" };
      }
    }

    // Handle link upload if a valid link is provided
    if (isValidLink(link)) {
      try {
        await sendLinkMessage(link, caption);
        return { msg: "Link upload successful" };
      } catch (linkError) {
        return { status: 500, msg: "Link upload failed" };
      }
    }

    // If neither files nor a valid link, return an error
    return { status: 400, msg: "Invalid file(s) or link" };
  } catch (err: any) {
    return {
      status: 500,
      msg: "An unexpected error occurred",
      error: err.message,
    };
  }
}

// Helper function to build the caption for the message
function buildCaption(name: string, link: string) {
  let caption = name ? `Uploaded by @${name}` : "";
  const { success } = urlSchema.safeParse(link);

  if (success) {
    caption =
      (link.includes("drive.google.com")
        ? `[Drive Link](${link})\n`
        : `[Web Link](${link})\n`) + caption;
  }

  return caption;
}

// Helper function to check if the link is valid
function isValidLink(link: string) {
  const { success } = urlSchema.safeParse(link);
  return success;
}

// Handles the upload of a single or multiple files
async function handleFileUpload(files: File[], caption: string) {
  const file = isValidFile(files[0]);
  const filename = nanoid(6);

  if (files.length === 1 && file) {
    return await uploadSingleFile(file, caption, filename);
  } else if (files.length > 1) {
    return await uploadMultipleFiles(files, caption, filename);
  }

  return { status: 400, msg: "Invalid file(s)" };
}

// Helper function to upload a single file
async function uploadSingleFile(file: File, caption: string, filename: string) {
  const buf = await file.arrayBuffer();
  const filebuf = Buffer.from(buf);
  await bot.sendDocument(
    chatId,
    filebuf,
    {
      caption,
      parse_mode: "MarkdownV2",
    },
    { filename },
  );
  return { msg: "file(1) upload successful" };
}

// Helper function to zip and upload multiple files
async function uploadMultipleFiles(
  files: File[],
  caption: string,
  filename: string,
) {
  const { zipBuffer, count } = await zipper(files);
  console.log(`ZIP buffer size: ${zipBuffer.length} bytes`);
  await bot.sendDocument(
    chatId,
    zipBuffer,
    {
      caption,
      parse_mode: "MarkdownV2",
    },
    { filename },
  );
  return { msg: `file(${count}) upload successful` };
}

// Helper function to send a link message
async function sendLinkMessage(link: string, caption: string) {
  await bot.sendMessage(chatId, caption, { parse_mode: "MarkdownV2" });
}

async function zipper(files: File[]) {
  const zip = new AdmZip();
  let count = 0;

  for (const file of files) {
    if (isValidFile(file)) {
      const buffer = await file.arrayBuffer();
      zip.addFile(file.name, Buffer.from(buffer));
      count++;
    } else {
      console.warn(`Invalid file skipped: ${file.name}`);
    }
  }

  // Return the ZIP buffer

  return { zipBuffer: zip.toBuffer(), count };
}

function isValidFile(obj: any) {
  // Define valid image and PDF MIME types
  const validTypes = [
    "image/jpeg",
    "image/png",
    "image/gif",
    "application/pdf",
  ];

  // Check if the file type is one of the valid types
  return obj instanceof File &&
    obj.size < MAX_FILE_SIZE &&
    validTypes.includes(obj.type)
    ? obj
    : null;
}
