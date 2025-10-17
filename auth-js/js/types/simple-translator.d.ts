declare module '@andreasremdt/simple-translator' {
  export interface TranslatorOptions {
    filesLocation?: string;
    defaultLanguage?: string;
    detectLanguage?: boolean;
    selector?: string;
    debug?: boolean;
    registerGlobally?: string | boolean;
    persist?: boolean;
    persistKey?: string;
  }

  export class Translator {
    constructor(options?: TranslatorOptions);
    
    // Core translation methods
    translatePageTo(language?: string): void;
    translateForKey(key: string, language?: string): string | null;
    
    // Language management
    add(language: string, translation: Record<string, any>): Translator;
    remove(language: string): Translator;
    fetch(languageFiles: string | string[], save?: boolean): Promise<any>;
    
    // Properties
    get currentLanguage(): string;
    get defaultLanguage(): string;
    setDefaultLanguage(language: string): void;
  }

  export default Translator;
}
