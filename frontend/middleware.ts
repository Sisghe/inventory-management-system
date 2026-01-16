import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  // proteggiamo tutto ciò che sta sotto /dashboard
  if (pathname.startsWith("/dashboard")) {
    const token = req.cookies.get("access_token")?.value;

    // se non abbiamo token in cookie -> login
    if (!token) {
      const url = req.nextUrl.clone();
      url.pathname = "/login";
      return NextResponse.redirect(url);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*"],
};
