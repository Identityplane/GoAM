import LoginPage from "../../../../components/login-page";
import { connection } from "next/server";

export default async function DynamicAuthPage() {
  // wait for an incoming request to render this page
  await connection();

  return <LoginPage />;
}
