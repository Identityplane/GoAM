import React from 'react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';

interface DebugInspectorProps {
  debugData: any;
  onClose: () => void;
}

/**
 * A professional syntax highlighter for JSON with a soft dark theme
 */
const SyntaxHighlightedJson = ({ data }: { data: any }) => {
  if (!data) return null;
  
  const jsonString = JSON.stringify(data, null, 2);
  
  // Refined regex for dark theme syntax highlighting
  const highlighted = jsonString.replace(
    /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
    (match) => {
      let cls = 'text-amber-400'; // numbers
      if (/^"/.test(match)) {
        if (/:$/.test(match)) {
          cls = 'text-sky-400 font-medium'; // keys
        } else {
          cls = 'text-teal-400'; // strings
        }
      } else if (/true|false/.test(match)) {
        cls = 'text-orange-400'; // booleans
      } else if (/null/.test(match)) {
        cls = 'text-slate-500'; // null
      }
      return `<span class="${cls}">${match}</span>`;
    }
  );

  return (
    <pre 
      className="text-[11px] leading-relaxed font-mono text-slate-300 p-6 rounded-xl bg-slate-950/50 border border-slate-800/50 shadow-inner overflow-x-auto whitespace-pre-wrap selection:bg-sky-500/20"
      dangerouslySetInnerHTML={{ __html: highlighted }}
    />
  );
};

export function DebugInspector({ debugData, onClose }: DebugInspectorProps) {
  return (
    <div className="hidden lg:flex flex-col w-1/3 border-l border-slate-800 bg-slate-900 h-screen transition-all duration-300 animate-in slide-in-from-right shrink-0">
      <div className="p-5 border-b border-slate-800 flex justify-between items-center bg-slate-900/80 backdrop-blur-sm h-[64px] shrink-0 sticky top-0 z-10">
        <div>
          <h3 className="font-bold text-sm text-slate-100 tracking-tight">Debug Inspector</h3>
          <p className="text-[9px] text-slate-500 uppercase tracking-widest font-bold">Real-time Session State</p>
        </div>
        <button 
          onClick={onClose}
          className="p-1.5 hover:bg-slate-800 text-slate-400 hover:text-slate-100 rounded-lg transition-all duration-200 active:scale-95"
          aria-label="Close inspector"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M18 6 6 18" /><path d="m6 6 12 12" /></svg>
        </button>
      </div>
      <div className="flex-1 overflow-hidden bg-slate-900">
        <ScrollArea className="size-full">
          <div className="p-6">
            {debugData ? (
              <SyntaxHighlightedJson data={debugData} />
            ) : (
              <div className="flex flex-col items-center justify-center h-64 text-slate-500 italic text-center gap-3">
                <div className="p-4 rounded-full bg-slate-800/50">
                  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-slate-600"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                </div>
                <div className="space-y-1">
                  <p className="text-sm font-medium not-italic text-slate-400">Waiting for Data</p>
                  <p className="text-[10px] non-italic">Start an authentication flow to capture state.</p>
                </div>
              </div>
            )}
          </div>
        </ScrollArea>
      </div>
    </div>
  );
}
