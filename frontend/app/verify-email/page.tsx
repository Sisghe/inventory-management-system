import VerifyEmailClient from "./VerifyEmailClient";

type PageProps = {
  searchParams: { token?: string };
};

export default function VerifyEmailPage({ searchParams }: PageProps) {
  const token = searchParams?.token ?? "";
  return <VerifyEmailClient token={token} />;
}
