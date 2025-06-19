import prisma from "./db";

import {
  avPqs,
  avCourses,
  faculty,
  avLevels,
  avSessions,
  ceg510,
  engrDepts,
  ceg22,
} from "../src/lib/data";
import { ceg32 } from "../src/lib/data";

async function uploadrealDepts() {
  // const faculty = await createFaculty();
  console.log("Begin Creating Departments");
  for (const [id, title] of Object.entries(engrDepts)) {
    const dept = await prisma.department.create({
      data: {
        id,
        title,
        facultyId: 1,
      },
    });
  }

  console.log("End Creating Departments");
}
async function uploadRealCourses() {
  console.log("Begin Creating Courses");
  const res: any = [];
  for (const course of ceg22) {
    // const deptId = getCourseDeptId(course.id);
    const { departmentIDs, ...c } = course;
    await prisma.course.create({
      data: { ...c },
    });
    let newCourse;
    for (const deptId of departmentIDs) {
      newCourse = await prisma.course.update({
        where: { id: course.id },
        data: { departments: { connect: { id: deptId } } },
      });
    }
    console.log(newCourse);
    res.push(newCourse);
  }
  console.log("End Creating Courses");
  return res;
}
function getCourseDeptId(courseId: string) {
  return courseId.slice(0, 3);
}
async function uploadRealQuestions() {
  console.log("Begin Creating Questions");
  const res = await prisma.question.createMany({ data: avPqs });
  console.log("End Creating Questions");
}
async function uploadRealLevels() {
  // Generate levels
  console.log("Begin creating levels");
  const data = avLevels.map((l) => ({
    id: Number(l),
  }));
  await prisma.level.createMany({ data });
  console.log("End creating levels");
}

async function createFaculty(id: number) {
  console.log("Begin creating faculties");
  const data = await prisma.faculty.create({
    data: {
      id,
      title: faculty,
    },
  });
  console.log("End creating faculties");
  return data;
}

async function uploadRealSessions() {
  const data = avSessions.map((s) => ({ id: s }));
  await prisma.session.createMany({ data });
}

async function change() {
  // const res = await prisma.faculty.delete({
  //   where: { title: "Engineering" },
  // });
  // console.log("Deleted", res);
  // await prisma.faculty.create({
  //   data: { id: 1, title: "Engineering" },
  // });
  // await prisma.department.update({
  //   where: { id: "CEG" },
  //   data: { facultyId: 1 },
  // });
  // const courses = ceg520.map((c) => c.id);

  // const courses = ceg520.map((c) => c.id);
  // const oldCourses = await prisma.course.findMany({
  //   where: { id: { in: ["BUS440", "CIL524", "CEG514", ...courses] } },
  // });

  // const depts = await prisma.level.update({
  //   where: { id: 500 },
  //   data: {
  //     courses: {
  //       connect: oldCourses.map((course) => ({ id: course.id })),
  //     },
  //   },
  // });
  const courses = await prisma.course.updateMany({
    where: { id: { in: ["BUS440", "CIL524"] } },
    data: { departmentIDs: { push: "BME" } },
  });
  const depts = await prisma.department.update({
    where: { id: "BME" },
    data: { courseIDs: ["BUS440", "CIL524"] },
  });
}

async function uploadCivilCourses() {
  console.log("Begin Creating Courses");
  const res: any = [];
  for (const course of ceg510) {
    let deptId = getCourseDeptId(course.id);
    if (["BUS", "CIL"].includes(deptId)) {
      deptId = "CEG";
    }
    const { id, ...data } = course;
    const newCourse = await prisma.course.upsert({
      where: { id: course.id },
      create: { ...course, departments: { connect: { id: deptId } } },
      update: { ...data, departments: { connect: { id: deptId } } },
    });
    console.log("New course created or updated", newCourse);
    res.push(newCourse);
  }
  console.log("End Creating Courses");
  return res;
}
(async function uploadRealData() {
  // await resetTables();
  // await uploadrealDepts();
  // await uploadRealLevels();
  await uploadRealCourses();
  // await uploadRealSessions();
  // await uploadRealQuestions();
  // await uploadCivilCourses();
  // await change();
})();
