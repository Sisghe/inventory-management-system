import VerifyEmailClient from "./VerifyEmailClient";

type PageProps = {
  searchParams: Promise<{ token?: string }>;
};

export default async function VerifyEmailPage({ searchParams }: PageProps) {
  const sp = await searchParams;
  const token = sp?.token ?? "";
  return <VerifyEmailClient token={token} />;
}
