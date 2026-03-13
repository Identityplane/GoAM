import { NextResponse } from 'next/server';
import fs from 'fs';
import path from 'path';

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const configName = searchParams.get('config') || 'identityplane';

  try {
    const configPath = path.join(process.cwd(), 'app/api/settings/configs', `${configName}.json`);
    const fileContents = fs.readFileSync(configPath, 'utf8');
    const config = JSON.parse(fileContents);
    return NextResponse.json(config);
  } catch (error) {
    // Fallback to default if config not found
    return NextResponse.json({
      backgroundColor: '#3F3FF3',
      accentColor: '#3F3FF3',
      logoSvg: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" fill="none"><rect width="32" height="32" rx="8" fill="currentColor"/><rect x="8" y="8" width="16" height="16" rx="4" fill="white"/></svg>',
      logoName: 'IdentityPlane',
      privacyPolicyUrl: '#',
      sidebarTitle: 'Effortlessly manage your team and operations.',
      sidebarText: 'Access your CRM dashboard and manage your team.',
      fontFamily: 'Inter, "Inter Fallback"',
      show_sidebar: true,
      pageBackgroundColor: '#ffffff',
      inputBackgroundColor: '#ffffff'
    });
  }
}
