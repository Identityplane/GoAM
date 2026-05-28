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
      logoSvg: '',
      logoName: '',
      privacyPolicyUrl: '#',
      sidebarTitle: '',
      sidebarText: '',
      fontFamily: '',
      show_sidebar: true,
      pageBackgroundColor: '#ffffff',
      inputBackgroundColor: '#ffffff'
    });
  }
}
