import { FormSchema, filterSchema } from "../types/client.d";
import { avDepts, engrDepts } from "../data";
import { getCookie } from "../actions";

export function formatParams(data: FormSchema) {
  function getKeyByValue(object: any, value: string) {
    return Object.keys(object).find((key) => object[key] === value) as string;
  }
  const deptId = getKeyByValue(engrDepts, data.department)?.toLowerCase();
  const levelId = data.level.split(" ")[0];
  const semester = data.semester.split(" ")[1];
  return { deptId, levelId, semester };
}
export async function formatCookies() {
  const courseFilter = await getCookie();
  let level = "500 Level";
  let semester = "Semester 2";
  let department = avDepts.get("CEG");
  if (courseFilter) {
    const { deptId, levelId, semester: semNumber } = courseFilter;
    department = avDepts.get(deptId.toUpperCase());
    level = `${levelId} Level`;
    semester = `Semester ${semNumber}`;
  }
  return { department, level, semester };
}
export function formatCourseDefaults(filter: string) {
  // let level = "500 Level";
  // let semester = "Semester 2";
  // let department = avDepts.get("CEG");
  let level, semester, department;
  const res = filterSchema.safeParse(filter ? JSON.parse(filter) : "");
  if (res.success) {
    const { deptId, levelId, semester: semNumber } = res.data;
    department = avDepts.get(deptId.toUpperCase());
    level = `${levelId} Level`;
    semester = `Semester ${semNumber}`;
  }
  return { department, level, semester };
}

export async function getCoursesUrl(qId: string) {
  const courseId = qId.split("-")[0];
  const deptId = courseId.slice(0, 3);
  const levelId = courseId.slice(3, 4) + "00";
  const semester = Number(courseId.slice(4, 5)) % 2 === 0 ? 2 : 1;
  return `/courses/${deptId}/${levelId}/${semester}`;
}
