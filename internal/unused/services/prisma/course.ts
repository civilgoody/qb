import db from "@/prisma/db";
import { CourseTable, TFilterSchema, filterSchema } from "../../types/client.d";
import { CourseStatus } from "@prisma/client";
import { notFound } from "next/navigation";
class PrismaCourseService {
  async getAll() {
    const data = await db.course.findMany({});
    return data;
  }
  async getByCourse(id: string) {
    const data = await db.course.findMany({
      where: { id },
    });
    return data;
  }
  async getCourseById(id: string) {
    const course = await db.course.findFirst({ where: { id } });
    if (!course) notFound();
    return course;
  }

  async mainFilter(filter: TFilterSchema) {
    let data: CourseTable[] = [];

    const { deptId, levelId, semester } = filter;
    data = await db.course.findMany({
      where: {
        departments: { some: { id: deptId.toUpperCase() } },
        levelId: Number(levelId),
        semester: Number(semester),
        status: { not: CourseStatus.UNAVAILABLE },
      },
      select: {
        id: true,
        units: true,
        title: true,
        levelId: true,
        semester: true,
        status: true,
        departmentIDs: true,
        questions: {
          select: {
            id: true,
          },
          orderBy: {
            sessionId: "desc",
          },
          where: {
            imageLinks: { isEmpty: false },
            type: "EXAM",
          },
        },
      },
      orderBy: {
        status: "asc",
      },
    });
    return { data };
  }
}

export function getCourseDetails(filter: string[]) {
  const res = filterSchema.safeParse({
    deptId: filter[0],
    levelId: filter[1],
    semester: filter[2],
  });
  if (!res.success) {
    notFound();
  }
  return res.data;
}

export const courseService = new PrismaCourseService();
