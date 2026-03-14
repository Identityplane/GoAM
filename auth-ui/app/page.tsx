import LoginPage from "@/components/login-page";
import { connection } from "next/server";

export default async function Page() {
  await connection();
  return <LoginPage />;
}
