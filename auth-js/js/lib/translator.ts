import Translator from '@andreasremdt/simple-translator';


export function initTranslator() {

    // Check if we have any link attribute in the head with rel="translation"
    let locales = new Map<string, string>();

    var translator = new Translator({debug: true});
    window.translator = translator;

    const links = document.head.querySelectorAll('link[i18n]');
    if (links.length > 0) {
        for (const link of links) {
            const href = (link as HTMLLinkElement).href;
            const locale = link.getAttribute('i18n') as string;
            locales.set(locale, href);
        }
    }

    console.log('locales', locales);
    translatePageWithLocaleCookie(translator, locales);
}

// Cache for translation files to avoid repeated fetches
const translationCache = new Map<string, any>();

async function translatePageWithLocaleCookie(translator: Translator, locales: Map<string, string>){

    // Get the value from the locale cookie
    const locale = document.cookie.split('; ').find(row => row.startsWith('locale='))?.split('=')[1];
    if (locale) {
        const localeFile = locales.get(locale);
        if (localeFile) {
            const data = await loadTranslationFile(locale, localeFile);
            translator.add(locale, data).translatePageTo(locale);
        }
    }
}

async function loadTranslationFile(locale: string, localeFile: string): Promise<any> {
    try {
        // Check if we already have this translation cached
        if (translationCache.has(locale)) {
            return translationCache.get(locale);
        }

        // Fetch with cache control headers
        const response = await fetch(localeFile, {
            cache: 'force-cache', // Use browser cache
        });

        if (!response.ok) {
            throw new Error(`Failed to fetch translation file: ${response.status} ${response.statusText}`);
        }

        return await response.json();
        
    } catch (error) {
        console.error('Error loading translation file:', error);
        // Fallback: try to load without caching
        try {
            const response = await fetch(localeFile);
            return await response.json();
        } catch (fallbackError) {
            console.error('Fallback translation loading also failed:', fallbackError);
        }
    }
}