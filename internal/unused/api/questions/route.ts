import z from "zod";
import prisma from "@/prisma/db";
import apiHandler from "@/src/utils/api/api.handler";
import * as Boom from "@hapi/boom";
import { NextResponse as res } from "next/server";
import { questionService } from "@/src/lib/services/prisma/question";
import { genQuestionId, parseSessionEnv } from "@/src/utils/helpers";
const EXISTS_ERR =
  "We are no longer accepting uploads for this " +
  "question at this time, try again later or contact support.";
export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const courseId = searchParams.get("courseId") ?? "";
  const courseWithQuestions = await questionService.getByCourse(courseId);
  return res.json(courseWithQuestions);
}
const DataSchema = z.object({
  courseId: z
    .string({ required_error: "Provide Course Id" })
    .min(6, "Provide a valid Course Id e.g ceg545")
    .transform((s) => s.slice(0, 6).toUpperCase()),
  sessionId: z
    .string({ required_error: "Provide Session Title" })
    .length(9, "Provide a valid session e.g 2023-2024")
    .transform((s) => s.replace("/", "-")),
  type: z.enum(["TEST", "EXAM"], {
    required_error: "Provide Question Type - TEST/EXAM",
  }),
});

type DataSchemaType = z.infer<typeof DataSchema>;

async function createQ(req: Request) {
  const validData = DataSchema.parse(await req.json());
  const data = await checkQ(validData);
  const q = await prisma.question.findFirst({
    where: { id: data.id },
    select: { id: true, imageLinks: true },
  });
  if (!q) {
    const result = await prisma.question.create({
      data,
      select: { id: true },
    });
    return res.json({ data: result }, { status: 201 });
  } else if (q.imageLinks.length < 3) {
    return res.json({
      data: { id: q.id },
      msg: "Provide Image Urls to complete upload",
    });
  } else throw Boom.conflict(EXISTS_ERR);
}
async function updateQ(req: Request) {
  const schema = z.object({
    id: z.string({ required_error: "Provide a valid Question Id" }),
    imgLink: z.string().min(5, "Provide a valid image url"),
  });
  const { id, imgLink } = schema.parse(await req.json());

  const q = await prisma.question.findFirst({
    where: { id },
    select: { id: true, imageLinks: true },
  });
  if (!q) {
    throw Boom.conflict(
      "Question doesn't exist, create via a POST request to `/api/questions` first",
    );
  } else if (q.imageLinks.length < 3) {
    const newImgLinks = [imgLink, ...q.imageLinks];
    const result = await prisma.question.update({
      where: { id },
      data: { imageLinks: newImgLinks },
      select: { id: true, imageLinks: true },
    });
    return res.json({ data: result }, { status: 201 });
  } else throw Boom.forbidden(EXISTS_ERR);
}
async function checkQ(data: DataSchemaType) {
  const { courseId, sessionId } = data;
  const sessions = parseSessionEnv();
  const id = genQuestionId(data);

  if (!sessions.includes(sessionId)) {
    throw Boom.forbidden(
      `Session not yet supported, We currently support ` +
        `sessions ${sessions[sessions.length - 1]} to ${sessions[0]}`,
    );
  }

  const course = await prisma.course.findFirst({
    where: { id: courseId },
  });
  const errMsg = "Your course or department is not yet supported";
  if (!course) throw Boom.forbidden(errMsg);
  return { id, courseId, sessionId, type: data.type };
}

export const POST = apiHandler({
  POST: createQ,
});
export const PUT = apiHandler({
  PUT: updateQ,
});
