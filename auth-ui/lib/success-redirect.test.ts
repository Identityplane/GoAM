import { getRedirectUrl } from './success-redirect';

describe('getRedirectUrl', () => {
  test.each([
    {
      name: 'no result',
      input: undefined,
      want: null,
    },
    {
      name: 'success false does not redirect',
      input: { success: false, redirect: 'https://example.com' },
      want: null,
    },
    {
      name: 'success true without redirect does not redirect',
      input: { success: true },
      want: null,
    },
    {
      name: 'success true with non-string redirect does not redirect',
      input: { success: true, redirect: 123 },
      want: null,
    },
    {
      name: 'success true with empty redirect does not redirect',
      input: { success: true, redirect: '   ' },
      want: null,
    },
    {
      name: 'success true with redirect returns trimmed url',
      input: { success: true, redirect: '  https://example.com/callback  ' },
      want: 'https://example.com/callback',
    },
  ])('$name', ({ input, want }) => {
    expect(getRedirectUrl(input as any)).toBe(want);
  });
});

