import { Course, CourseStatus } from "@prisma/client";
import prisma from "@/prisma/db";
import { z } from "zod";
import { NextResponse } from "next/server";
import apiHandler from "@/src/utils/api/api.handler";
import { revalidatePath } from "next/cache";
import { courseService } from "@/src/lib/services/prisma/course";

// Validation schema for a single course
const SingleCourseSchema = z.object({
  title: z.string().min(5, "Course title length must exceed 5 characters"),
  id: z
    .string()
    .min(6, "Course id must be exactly 6 characters e.g. CEG223")
    .max(6, "Course id must be exactly 6 characters e.g. CEG516")
    .toUpperCase(),
  units: z.enum(["1", "2", "3", "4"]).transform(Number),
  levelId: z.enum(["100", "200", "300", "400", "500"]).transform(Number),
  semester: z.enum(["1", "2"]).transform(Number),
  description: z.string().optional(),
  status: z.nativeEnum(CourseStatus).optional(),
  deptId: z.string().min(3, "Provide valid department ID").max(3).toUpperCase(),
});
type newCourseSchema = z.infer<typeof SingleCourseSchema>;

// Validation schema for multiple courses
const MultipleCourseUploadSchema = z.array(SingleCourseSchema);

async function upsertCourses(req: Request) {
  const coursesData = MultipleCourseUploadSchema.parse(await req.json());

  const processedResults = {
    created: [] as any[],
    updated: [] as any[],
    errors: [] as any[],
  };

  for (const course of coursesData) {
    try {
      const existingCourse = await prisma.course.findUnique({
        where: { id: course.id },
      });

      if (!existingCourse) {
        // If course doesn't exist, create it
        const newCourse = await prisma.course.create({
          data: {
            title: course.title,
            id: course.id,
            units: course.units,
            levelId: course.levelId,
            semester: course.semester,
            description: course.description ?? null,
            status: course.status ?? null,
            departments: { connect: { id: course.deptId } },
          },
        });
        processedResults.created.push(newCourse);
      } else if (isUnchanged(course, existingCourse)) {
        continue;
      } else {
        const mergedDepartmentIDs = Array.from(
          new Set([...existingCourse.departmentIDs, course.deptId]),
        );
        // If course exists, update it
        const updatedCourse = await prisma.course.update({
          where: { id: course.id },
          data: {
            title: course.title,
            units: course.units,
            levelId: course.levelId,
            semester: course.semester,
            description: course.description ?? null,
            status: course.status ?? null,
            departmentIDs: mergedDepartmentIDs,
          },
        });
        processedResults.updated.push(updatedCourse);
      }
    } catch (error: any) {
      processedResults.errors.push({
        courseId: course.id,
        error: error.message || "Error processing course",
      });
    }
  }
  if (
    processedResults.created.length > 0 ||
    processedResults.updated.length > 0
  ) {
    // const deptId = coursesData[0].departmentIDs[0];
    const { deptId, levelId, semester } = coursesData[0];
    revalidatePath(`/courses/${deptId}/${levelId}/${semester}`);
  }
  // Return a summary of processed results
  return NextResponse.json(
    {
      message: "Bulk course processing complete",
      results: processedResults,
    },
    { status: 200 },
  );
}
function isUnchanged(course: newCourseSchema, existingCourse: Course) {
  const isUnchanged =
    existingCourse.title === course.title &&
    existingCourse.units === course.units &&
    existingCourse.levelId === course.levelId &&
    existingCourse.semester === course.semester &&
    existingCourse.description === (course.description ?? null) &&
    existingCourse.status === (course.status ?? null) &&
    existingCourse.departmentIDs.includes(course.deptId);
  return isUnchanged;
}
export async function GET() {
  const courses = await courseService.mainFilter({
    deptId: "ceg",
    levelId: "500",
    semester: "1",
  });
  return NextResponse.json(courses);
}
export const POST = apiHandler({
  POST: upsertCourses,
});
export const PUT = apiHandler({
  PUT: upsertCourses,
});
