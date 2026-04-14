import QRCode from 'qrcode';

export async function encodeQrCodeDataUrl(text: string): Promise<string> {
  if (!text) {
    throw new Error('text is required');
  }

  return QRCode.toDataURL(text, {
    errorCorrectionLevel: 'M',
    margin: 1,
    width: 240,
  });
}

