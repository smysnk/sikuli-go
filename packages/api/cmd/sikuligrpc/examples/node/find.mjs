import { ensureSikuligoOnPath } from "./bootstrap.mjs";
import { Screen, Pattern } from "@sikuligo/sikuligo";

ensureSikuligoOnPath();

const screen = await Screen();
try {
  const pattern = Pattern("assets/pattern.png").exact();
  const match = await screen.find(pattern);
  console.log(`found rect=(${match.x},${match.y},${match.w},${match.h}) target=(${match.targetX},${match.targetY})`);
} finally {
  await screen.close();
}
