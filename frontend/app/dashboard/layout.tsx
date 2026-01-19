import DashboardNavbar from "./navbar";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-vh-100 d-flex flex-column">
      <DashboardNavbar />

      <main className="flex-grow-1">
        {/* container-xxl: più “professionale” su schermi grandi, ma resta leggibile su mobile */}
        <div className="container-xxl py-4 px-3">{children}</div>
      </main>
    </div>
  );
}
