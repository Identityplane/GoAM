import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { FieldGroup, Field } from '@/components/ui/field';
import { AuthStepProps } from './types';
import { useState, useEffect } from 'react';

export function TermsStep({ isLoading, onContinue, accentColor, formData, settings, error }: AuthStepProps) {
  const [agreedTc, setAgreedTc] = useState(false);
  const [agreedPrivacy, setAgreedPrivacy] = useState(false);
  const [agreedMarketing, setAgreedMarketing] = useState(false);

  useEffect(() => {
    if (settings) {
      if (settings.enable_tc_terms && settings.preselect_tc_terms) setAgreedTc(true);
      if (settings.enable_privacy_terms && settings.preselect_privacy_terms) setAgreedPrivacy(true);
      if (settings.enable_marketing_terms && settings.preselect_marketing_terms) setAgreedMarketing(true);
    }
  }, [settings]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onContinue({
      agreedTc,
      agreedPrivacy,
      agreedMarketing
    });
  };

  const isContinueDisabled = isLoading || 
    (settings?.enable_tc_terms && !agreedTc) || 
    (settings?.enable_privacy_terms && !agreedPrivacy);

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl text-foreground">Terms & Conditions</h2>
        <p className="text-muted-foreground">Please review and accept our policies to continue.</p>
      </div>


      <form onSubmit={handleSubmit} className="space-y-6">
        <FieldGroup>
          <div className="space-y-4">
            {settings?.enable_tc_terms && (
              <div className="flex items-start space-x-3 p-3 rounded-lg border border-gray-100 hover:bg-gray-50 transition-colors">
                <Checkbox
                  id="tc"
                  checked={agreedTc}
                  onCheckedChange={(checked) => setAgreedTc(checked as boolean)}
                  disabled={isLoading}
                  className="mt-1"
                />
                <div className="grid gap-1.5 leading-none">
                  <Label htmlFor="tc" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer">
                    I agree to the <a href={settings.termsOfConditionsUrl || "#"} target="_blank" rel="noopener noreferrer" className="text-primary hover:underline" style={{ color: accentColor }}>Terms of Conditions</a>
                  </Label>
                </div>
              </div>
            )}

            {settings?.enable_privacy_terms && (
              <div className="flex items-start space-x-3 p-3 rounded-lg border border-gray-100 hover:bg-gray-50 transition-colors">
                <Checkbox
                  id="privacy"
                  checked={agreedPrivacy}
                  onCheckedChange={(checked) => setAgreedPrivacy(checked as boolean)}
                  disabled={isLoading}
                  className="mt-1"
                />
                <div className="grid gap-1.5 leading-none">
                  <Label htmlFor="privacy" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer">
                    I have read and accept the <a href={settings.privacyPolicyUrl || "#"} target="_blank" rel="noopener noreferrer" className="text-primary hover:underline" style={{ color: accentColor }}>Privacy Policy</a>
                  </Label>
                </div>
              </div>
            )}

            {settings?.enable_marketing_terms && (
              <div className="flex items-start space-x-3 p-3 rounded-lg border border-gray-100 hover:bg-gray-50 transition-colors">
                <Checkbox
                  id="marketing"
                  checked={agreedMarketing}
                  onCheckedChange={(checked) => setAgreedMarketing(checked as boolean)}
                  disabled={isLoading}
                  className="mt-1"
                />
                <div className="grid gap-1.5 leading-none">
                  <Label htmlFor="marketing" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer">
                    I agree to receive marketing updates and newsletters
                  </Label>
                  <p className="text-xs text-muted-foreground">You can unsubscribe at any time.</p>
                </div>
              </div>
            )}
          </div>
        </FieldGroup>

        <Button type="submit" disabled={isContinueDisabled} className="w-full">
          {isLoading ? 'Processing...' : 'Accept & Continue'}
        </Button>
      </form>
    </div>
  );
}

