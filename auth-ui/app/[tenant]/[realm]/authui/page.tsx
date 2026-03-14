import LoginPage from "../../../../components/login-page";
import { connection } from "next/server";

export default async function DynamicAuthPage() {
  await connection();
  return <LoginPage />;
}
