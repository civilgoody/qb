import { Question } from "@prisma/client";

import prisma from "@/prisma/db";

import {
  engrDepts,
  courseLevels,
  avSessions,
  faculty,
  sessions,
} from "@/src/lib/data";
import { getRandItem, genQuestionId } from "@/src/utils/helpers";
import { faker } from "@faker-js/faker";
import { Course } from "@prisma/client";
import { QuestionType } from "@prisma/client";

(async function main() {
  try {
    console.log("Begin creating faculties");
    const newFaculty = await prisma.faculty.create({
      data: {
        id: 1,
        title: "Engineering",
        id: 1,
      },
    });
    console.log("End creating faculties");

    // Generate departments
    console.log("Begin creating departments");
    const deptData = Object.entries(engrDepts).map(([id, title]) => ({
      facultyId: newFaculty.id,
      id,
      title,
    }));

    await prisma.department.createMany({
      data: deptData,
      // skipDuplicates: , // Avoids duplicate entries
    });
    console.log("End creating departments");

    // Generate sessions
    console.log("Begin creating sessions");
    const sessionsData = sessions.map((session) => ({ id: session }));
    await prisma.session.createMany({
      data: sessionsData,
      // skipDuplicates: true,
    });

    console.log("End creating sessions");

    // Generate levels
    const levels = courseLevels.map((l) => ({ id: l }));
    await prisma.level.createMany({
      data: levels,
      // skipDuplicates: true,
    });

    // Generate courses
    console.log("Begin creating courses");
    const coursesData: Course[] = [];
    const generatedIds = new Set<string>();

    for (let i = 0; i < 200; i++) {
      const deptId = getRandItem(deptData).id;
      const level = getRandItem(courseLevels);
      let idString;

      do {
        idString =
          `${deptId}${level.toString().charAt(0)}${faker.number.int({ min: 11, max: 90 })}`.toUpperCase();
      } while (generatedIds.has(idString));
      generatedIds.add(idString);

      const courseData = await prisma.course.create({
        data: {
          id: idString.toUpperCase(),
          units: faker.number.int({ min: 1, max: 3 }),
          title: faker.commerce.productName(),
          levelId: level,
          semester: getRandItem([1, 2]),
          description: null,
          createdAt: new Date(),
          updatedAt: new Date(),
          departments: {
            connect: {
              id: deptId,
            },
          },
        },
      });
      coursesData.push(courseData);
      console.log(`Created new Course ${i + 1}`);
    }
    console.log("End creating courses");

    // Generate questions
    console.log("Begin Creating Questions");
    const questionsData: Question[] = [];
    const generatedQIds = new Set<string>();

    for (let i = 0; i < 600; i++) {
      const courseId = getRandItem(coursesData)?.id;
      const sessionId = getRandItem(sessionsData)?.id;
      const type = getRandItem(["TEST", "EXAM"]) as QuestionType;
      let idString;

      do {
        idString = genQuestionId({ courseId, sessionId, type });
        console.log("New Question Id", idString);
      } while (generatedQIds.has(idString));
      generatedQIds.add(idString);

      const question = await prisma.question.create({
        data: {
          id: idString,
          imageLinks: [faker.image.url()],
          courseId: courseId,
          sessionId: sessionId,
          lecturer: faker.person.fullName(),
          type: type,
          docLink: null,
          timeAllowed: null,
          tips: null,
          downloads: null,
          views: null,
          approved: getRandItem([false, true]),
          uploaderId: null,
          createdAt: new Date(),
          updatedAt: new Date(),
        },
        // skipDuplicates: true, // Avoids duplicate entries
      });
      questionsData.push(question);
      console.log(`Created new Question ${i + 1}`);
    }
    console.log("End Creating Questions");

    console.log("Database populated successfully!");
  } catch (error) {
    console.error("Error generating data:", error);
  } finally {
    await prisma.$disconnect();
  }
})();

// const goodTables = ["",  "_prisma_migrations"]
async function resetTables() {
  try {
    console.log("Begin Resetting Tables");

    const delLevels = prisma.level.deleteMany({});
    const delQuestions = prisma.question.deleteMany({});
    const delDepts = prisma.department.deleteMany({});
    const delCourses = prisma.course.deleteMany({});
    const delFaculties = prisma.faculty.deleteMany({});
    const delUsers = prisma.user.deleteMany({});
    await prisma.$transaction([
      delLevels,
      delQuestions,
      delDepts,
      delCourses,
      delFaculties,
      delUsers,
    ]);
  } catch (error) {
    console.log({ error });
  } finally {
    console.log("End Resetting Tables");
  }
}

async function createFaculty() {
  console.log("Begin creating faculties");
  const data = await prisma.faculty.create({
    data: {
      title: faculty,
      id: 1,
    },
  });
  console.log("End creating faculties");
  return data;
}

async function createFakeData() {
  // Generate sessions
  console.log("Begin creating sessions");
  const sessionsData = avSessions.map((session) => ({ id: session }));
  await prisma.session.createMany({
    data: sessionsData,
    // skipDuplicates: true,
  });
  console.log("End creating sessions");

  // Generate departments
  console.log("Begin creating departments");
  const faculty = await createFaculty();
  const deptData = Object.entries(engrDepts).map(([id, title]) => ({
    facultyId: faculty.id,
    id,
    title,
  }));
  await prisma.department.createMany({
    data: deptData,
    // skipDuplicates: true, // Avoids duplicate entries
  });
  console.log("End creating departments");

  // Generate levels
  const levels = courseLevels.map((l) => ({ id: l }));
  await prisma.level.createMany({
    data: levels,
    // skipDuplicates: true,
  });
}
