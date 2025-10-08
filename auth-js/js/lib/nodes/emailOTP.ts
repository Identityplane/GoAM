import './emailOTP.css';

interface OTPInput extends HTMLInputElement {
  value: string;
}

interface OTPForm extends HTMLFormElement {
  submit(): void;
}

export function initEmailOTP(): void {
  const form: OTPForm | null = document.getElementById('otpForm') as OTPForm;
  const basicInput: OTPInput | null = document.getElementById('otp') as OTPInput;

  if (!form || !basicInput) {
    console.error('Required form elements not found');
    return;
  }

  initResendCountdown();

  // Create enhanced version
  const otpInputs = document.createElement('div');
  otpInputs.className = 'otp-inputs';
  
  // Create 6 input fields
  for (let i = 0; i < 6; i++) {
    const input = document.createElement('input');
    input.type = 'text';
    input.maxLength = 1;
    input.pattern = '[0-9]';
    input.inputMode = 'numeric';
    input.className = 'otp-input';
    input.required = true;
    otpInputs.appendChild(input);
  }

  // Insert enhanced version before the basic input
  basicInput.parentElement?.insertBefore(otpInputs, basicInput);

  // Add enhanced class to show the multi-input version
  otpInputs.classList.add('enhanced');

  const inputs: NodeListOf<OTPInput> = otpInputs.querySelectorAll('.otp-input');

  inputs.forEach((input: OTPInput, index: number) => {
    input.addEventListener('input', function(this: OTPInput) {
      if (this.value.length === 1) {
        if (index < inputs.length - 1) {
          inputs[index + 1].focus();
        }
      }
    });

    input.addEventListener('keydown', function(this: OTPInput, e: KeyboardEvent) {
      if (e.key === 'Backspace' && !this.value && index > 0) {
        inputs[index - 1].focus();
      }
    });

    input.addEventListener('paste', function(this: OTPInput, e: ClipboardEvent) {
      e.preventDefault();
      const pastedData = e.clipboardData?.getData('text').slice(0, 6) || '';
      if (/^\d+$/.test(pastedData)) {
        pastedData.split('').forEach((digit: string, i: number) => {
          if (inputs[i]) {
            inputs[i].value = digit;
          }
        });
        updateBasicInput();
        if (pastedData.length === 6) {
          form.submit();
        }
      }
    });
  });

  function updateBasicInput(): void {
    if (!basicInput) return;
    const otp = Array.from(inputs).map(input => input.value).join('');
    basicInput.value = otp;
  }

  form.addEventListener('submit', function(e: Event) {
    e.preventDefault();
    updateBasicInput();
    if (basicInput.value.length === 6) {
      form.submit();
    }
  });

  inputs.forEach((input: OTPInput) => {
    input.addEventListener('input', function() {
      updateBasicInput();
      if (basicInput.value.length === 6) {
        form.submit();
      }
    });
  });
}

// Helper functions for testing
export function getOTPValue(inputs: NodeListOf<OTPInput>): string {
  return Array.from(inputs).map(input => input.value).join('');
}

export function simulateInput(input: OTPInput, value: string): void {
  input.value = value;
  input.dispatchEvent(new Event('input'));
}

export function simulatePaste(input: OTPInput, value: string): void {
  const clipboardData = new DataTransfer();
  clipboardData.setData('text', value);
  const pasteEvent = new ClipboardEvent('paste', {
    clipboardData,
    bubbles: true,
    cancelable: true
  });
  input.dispatchEvent(pasteEvent);
} 

function initResendCountdown(): void {
  const resendButton = document.getElementById('resend-otp-button') as HTMLButtonElement;
  
  if (!resendButton) {
    return; // Button doesn't exist or countdown not needed
  }

  const resendInSeconds = parseInt(resendButton.getAttribute('data-resend-in-seconds') || '0');
  
  if (resendInSeconds <= 0) {
    return; // No countdown needed
  }

  let remainingSeconds = resendInSeconds;
  
  function updateButtonText(): void {
    if (remainingSeconds > 0) {
      resendButton.textContent = `Resend in (${remainingSeconds})`;
      resendButton.disabled = true;
      remainingSeconds--;
    } else {
      resendButton.textContent = 'Resend';
      resendButton.disabled = false;
      return; // Stop the countdown
    }
  }

  // Initial update
  updateButtonText();
  
  // Update every second
  const countdownInterval = setInterval(() => {
    updateButtonText();
    if (remainingSeconds < 0) {
      clearInterval(countdownInterval);
    }
  }, 1000);
}