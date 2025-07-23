export async function initTelegramLogin(): Promise<void> {

    console.log('initializing TelegramLogin')

    // Check if there is a tgAuthResult in the url fragment
    const tgAuthResult = window.location.hash.split('tgAuthResult=')[1]
    if (tgAuthResult) {

        console.log('tgAuthResult', tgAuthResult)

        // Set it to the form input
        const tgAuthResultInput = document.getElementById('tgAuthResult') as HTMLInputElement | null;
        if (tgAuthResultInput) {
            tgAuthResultInput.value = tgAuthResult;
        } else {
            console.error("Telegram auth result input not found");
            return;
        }

        // Submit the form
        const form = document.getElementById('telegramLoginForm') as HTMLFormElement | null;
        if (form) {
            form.submit();
        } else {
            console.error("Telegram login form not found");
        }
    }
}