import ResetPasswordClient from "./ResetPasswordClient";

type PageProps = {
  searchParams: Promise<{ token?: string }>;
};

export default async function ResetPasswordPage({ searchParams }: PageProps) {
  const sp = await searchParams;
  const token = sp?.token ?? "";
  return <ResetPasswordClient token={token} />;
}
