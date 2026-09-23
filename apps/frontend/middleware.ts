import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export async function middleware(request: NextRequest) {
  const isDashboardPage = request.nextUrl.pathname.startsWith('/dashboard');
  const isLandingPage = request.nextUrl.pathname === '/';
  if (!isDashboardPage && !isLandingPage) {
    return NextResponse.next();
  }

  // Fast check: if no session cookie, they definitely aren't authenticated
  const hasSessionCookie = request.cookies.has('session');

  if (!hasSessionCookie) {
    if (isDashboardPage) {
      return NextResponse.redirect(new URL('/', request.url));
    }
    return NextResponse.next();
  }

  if (isLandingPage) {
    return NextResponse.redirect(new URL('/dashboard', request.url));
  }


  return NextResponse.next();
}

export const config = {
  matcher: ['/', '/dashboard/:path*'],
};
