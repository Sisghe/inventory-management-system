import ResetPasswordClient from "./ResetPasswordClient";

type PageProps = {
  searchParams: { token?: string };
};

export default function ResetPasswordPage({ searchParams }: PageProps) {
  const token = searchParams?.token ?? "";
  return <ResetPasswordClient token={token} />;
}
