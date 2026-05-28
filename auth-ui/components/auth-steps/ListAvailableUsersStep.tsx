'use client';

import { Button } from '@/components/ui/button';
import { AuthStepProps } from './types';
import { Badge } from '@/components/ui/badge';
import { PlusCircle, LogIn, Clock } from 'lucide-react';

interface AvailableUser {
  userid: string;
  username?: string;
  email?: string;
  loa: number;
}

export function ListAvailableUsersStep({ isLoading, onContinue, accentColor, prompts }: AuthStepProps) {
  const usersJson = prompts?.users || '[]';
  let users: AvailableUser[] = [];

  try {
    users = JSON.parse(usersJson);
  } catch (e) {
    console.error('Failed to parse available users', e);
  }

  const getDisplayName = (user: AvailableUser) => {
    return user.email || user.username || user.userid;
  };

  const getInitials = (name: string) => {
    return name.slice(0, 2).toUpperCase();
  };

  const colorStyle = { color: accentColor || '#3F3FF3' };
  const bgStyle = { backgroundColor: accentColor || '#3F3FF3' };

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl text-foreground">Choose an account</h2>
        <p className="text-muted-foreground">Select an account to continue with.</p>
      </div>

      <div className="space-y-3">
        {users.map((user) => {
          const displayName = getDisplayName(user);
          const isLoggedOut = user.loa === 0;

          return (
            <button
              key={user.userid}
              disabled={isLoading}
              onClick={() => onContinue({ userid: user.userid })}
              className="w-full flex items-center p-4 rounded-xl border-2 border-muted hover:border-primary transition-all text-left group bg-card"
            >
              <div
                className="w-12 h-12 rounded-full flex items-center justify-center text-white font-bold text-lg mr-4 shrink-0 shadow-sm"
                style={bgStyle}
              >
                {getInitials(displayName)}
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <span className="font-bold text-foreground truncate">{displayName}</span>
                  {isLoggedOut && (
                    <span className="text-xs text-muted-foreground font-medium flex items-center gap-1 shrink-0">
                      <Clock size={12} />
                      Session expired
                    </span>
                  )}
                </div>
                {user.email && (user.username || user.userid !== user.email) && (
                  <div className="text-sm text-muted-foreground truncate">
                    {user.username || user.userid}
                  </div>
                )}
              </div>

              <LogIn className="ml-2 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" size={18} />
            </button>
          );
        })}

        <button
          disabled={isLoading}
          onClick={() => onContinue({ action: 'another' })}
          className="w-full flex items-center p-4 rounded-xl border border-dashed border-muted hover:border-primary/50 hover:bg-accent/5 transition-all text-left"
        >
          <div className="w-12 h-12 rounded-full border border-dashed border-muted flex items-center justify-center text-muted-foreground mr-4 shrink-0">
            <PlusCircle size={24} />
          </div>
          <span className="font-medium text-foreground">Use another account</span>
        </button>
      </div>
    </div>
  );
}
