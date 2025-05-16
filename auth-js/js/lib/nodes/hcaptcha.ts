import { hCaptchaLoader } from '@hcaptcha/loader';

export async function initHcaptcha(): Promise<void> {
    console.log('initializing Hcaptcha');
    
    await hCaptchaLoader();

    await hCaptchaLoader();

    hcaptcha.render({
    sitekey: '<your_sitekey>'
    });

    const { response } = await hcaptcha.execute({ async: true });

    console.log('hcaptcha response', response);

    const tokenInput = document.getElementById('hcaptcha-token') as HTMLInputElement;
    tokenInput.value = response;

    const form = document.getElementById('hcaptcha-form') as HTMLFormElement;
    form.submit();
}