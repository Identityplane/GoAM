import QRCode from 'qrcode';
import { encodeQrCodeDataUrl } from './qr-code';

jest.mock('qrcode', () => ({
  __esModule: true,
  default: {
    toDataURL: jest.fn(),
  },
}));

const mockedQRCode = QRCode as unknown as { toDataURL: jest.Mock };

describe('encodeQrCodeDataUrl', () => {
  it('calls qrcode.toDataURL with expected options', async () => {
    const cases = [
      { name: 'simple url', text: 'https://example.com', expected: 'data:image/png;base64,abc' },
      { name: 'login url', text: 'https://login.local/?secondary=secret', expected: 'data:image/png;base64,def' },
    ];

    for (const tc of cases) {
      mockedQRCode.toDataURL.mockResolvedValueOnce(tc.expected);
      await expect(encodeQrCodeDataUrl(tc.text)).resolves.toBe(tc.expected);

      expect(mockedQRCode.toDataURL).toHaveBeenLastCalledWith(
        tc.text,
        expect.objectContaining({
          errorCorrectionLevel: 'M',
          margin: 1,
          width: 240,
        })
      );
    }
  });

  it('throws on empty text', async () => {
    await expect(encodeQrCodeDataUrl('')).rejects.toThrow('text is required');
  });
});

