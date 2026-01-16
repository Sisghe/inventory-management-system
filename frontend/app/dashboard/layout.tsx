import DashboardNavbar from "./navbar";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <DashboardNavbar />
      <main className="container my-4">{children}</main>
    </>
  );
}
