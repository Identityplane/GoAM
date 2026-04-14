'use client';

import { useEffect, useMemo, useState } from 'react';
import { AuthStepProps } from './types';
import { encodeQrCodeDataUrl } from '@/lib/qr-code';
import { Button } from '@/components/ui/button';

export function QrCodeStep({ prompts, onRestart }: AuthStepProps) {
  const qrUrl = prompts?.qr_code || '';
  const [qrDataUrl, setQrDataUrl] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const printableUrl = useMemo(() => {
    if (!qrUrl) return '';
    // Avoid overly long lines in UI; the full URL is still available via copy.
    if (qrUrl.length <= 80) return qrUrl;
    return `${qrUrl.slice(0, 50)}…${qrUrl.slice(-20)}`;
  }, [qrUrl]);

  useEffect(() => {
    let cancelled = false;
    setQrDataUrl(null);
    setError(null);

    if (!qrUrl) {
      setError('QR code URL missing.');
      return;
    }

    encodeQrCodeDataUrl(qrUrl)
      .then((dataUrl) => {
        if (cancelled) return;
        setQrDataUrl(dataUrl);
      })
      .catch((e: any) => {
        if (cancelled) return;
        setError(e?.message || 'Failed to encode QR code.');
      });

    return () => {
      cancelled = true;
    };
  }, [qrUrl]);

  return (
    <div className="space-y-6 text-center">
      <div className="space-y-2 text-center mb-4">
        <h2 className="text-3xl text-foreground">Scan to sign in</h2>
        <p className="text-muted-foreground">
          Open the app on your logged-in phone and scan this QR code.
        </p>
      </div>

      <div className="flex justify-center">
        <div className="rounded-xl border bg-white p-4 shadow-sm">
          {qrDataUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={qrDataUrl}
              alt="Login QR code"
              className="h-60 w-60"
            />
          ) : (
            <div className="h-60 w-60 flex items-center justify-center text-sm text-muted-foreground">
              {error ? error : 'Generating QR code…'}
            </div>
          )}
        </div>
      </div>

      {qrUrl && (
        <div className="space-y-2">
          <div className="text-xs text-muted-foreground">QR URL</div>
          <div className="font-mono text-xs break-all text-muted-foreground">{printableUrl}</div>
          <Button
            type="button"
            variant="outline"
            onClick={() => navigator.clipboard?.writeText(qrUrl)}
          >
            Copy link
          </Button>
        </div>
      )}

      <div className="pt-2">
        <Button onClick={onRestart} className="w-full" variant="ghost">
          Start over
        </Button>
      </div>
    </div>
  );
}

