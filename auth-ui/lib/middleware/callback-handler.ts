import { NextRequest, NextResponse } from 'next/server';

/**
 * Handles POST requests to the authentication callback URL.
 * It parses the form data and redirects to the same URL using GET,
 * appending all form fields as query parameters.
 */
export async function handlePOSTCallback(request: NextRequest) {
  if (request.method !== 'POST') {
    return null;
  }

  const url = new URL(request.url);
  if (!url.pathname.endsWith('/callback')) {
    return null;
  }

  try {
    const formData = await request.formData();
    const searchParams = new URL(request.url).searchParams;

    // Add form data to search params
    formData.forEach((value, key) => {
      if (typeof value === 'string') {
        searchParams.append(key, value);
      }
    });

    // Redirect to the same URL but with GET and all parameters
    const getUrl = new URL(request.url);
    getUrl.search = searchParams.toString();

    return NextResponse.redirect(getUrl, 303);
  } catch (error) {
    console.error('Failed to parse POST callback body:', error);
    return null;
  }
}
