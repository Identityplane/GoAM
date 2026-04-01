import React from 'react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';

interface DebugInspectorProps {
  debugData: any;
  onClose: () => void;
  width: number;
  onWidthChange: (width: number) => void;
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

export function DebugInspector({ debugData, onClose, width, onWidthChange }: DebugInspectorProps) {
  const [isResizing, setIsResizing] = React.useState(false);
  const [isMounted, setIsMounted] = React.useState(false);

  React.useEffect(() => {
    setIsMounted(true);
  }, []);

  React.useEffect(() => {
    if (!isResizing) return;

    const handleMouseMove = (e: MouseEvent) => {
      // Calculate width from the right side
      const newWidth = window.innerWidth - e.clientX;
      
      // Constrain width
      if (newWidth > 300 && newWidth < window.innerWidth * 0.8) {
        onWidthChange(newWidth);
      }
    };

    const stopResizing = () => {
      setIsResizing(false);
      document.body.style.cursor = 'default';
      document.body.style.userSelect = 'auto';
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', stopResizing);
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', stopResizing);
    };
  }, [isResizing, onWidthChange]);

  return (
    <div 
      className={cn(
        "hidden lg:flex flex-col border-l border-slate-800 bg-slate-900 h-screen shrink-0 relative shadow-2xl z-50",
        !isResizing && "transition-all duration-300",
        !isMounted && "animate-in slide-in-from-right duration-500"
      )}
      style={{ width: `${width}px` }}
    >
      {/* Resize Handle - Wider hit area but skinny visual */}
      <div
        className="absolute -left-1.5 top-0 bottom-0 w-3 cursor-col-resize hover:bg-sky-500/20 active:bg-sky-500/40 transition-colors z-[100] group flex items-center justify-center"
        onMouseDown={() => setIsResizing(true)}
      >
        <div className="w-[2px] h-full bg-slate-800 group-hover:bg-sky-500/50 transition-colors pointer-events-none" />
        <div className="absolute top-1/2 -translate-y-1/2 w-4 h-12 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none">
          <div className="w-1 h-6 bg-slate-600 rounded-full" />
        </div>
      </div>

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
