import db from "@/prisma/db";
import { CourseFilter } from "@/src/lib/types/client";
import { redirect } from "next/navigation";
class PrismaQuestionService {
  async getAll() {
    const data = await db.question.findMany({});
    return data;
  }
  async getByCourse(courseId: string) {
    const course = await db.course.findFirst({
      select: {
        title: true,
        id: true,
        questions: {
          select: {
            id: true,
            sessionId: true,
            imageLinks: true,
            lecturer: true,
            type: true,
          },
          where: {
            imageLinks: { isEmpty: false },
          },
          orderBy: { sessionId: "desc" },
        },
      },
      where: {
        id: courseId.toUpperCase(),
      },
    });
    if (!course) {
      redirect("/not-found");
    }
    return { course };
  }

  async mainFilter({ deptId, levelId, semester }: CourseFilter) {
    const filteredData = await db.course.findMany({
      where: {
        departments: { some: { id: deptId.toUpperCase() } },
        levelId: Number(levelId),
        semester: Number(semester),
      },
      select: {
        id: true,
        units: true,
        title: true,
        levelId: true,
        semester: true,
        questions: {
          select: {
            id: true,
            imageLinks: true,
            sessionId: true,
            lecturer: true,
            type: true,
          },
        },
      },
    });
    return filteredData;
  }
}

export const questionService = new PrismaQuestionService();
