"use client";

import { Dialog, Transition } from "@headlessui/react";
import { Cancel01Icon, Alert01Icon } from "hugeicons-react";

interface ConfirmModalProps {
  isOpen: boolean;
  title: string;
  description: string;
  confirmText: string;
  onClose: () => void;
  onConfirm: () => void;
  isLoading: boolean;
  isDestructive?: boolean;
}

export function ConfirmModal({ 
  isOpen, 
  title, 
  description, 
  confirmText, 
  onClose, 
  onConfirm, 
  isLoading,
  isDestructive = true
}: ConfirmModalProps) {
  return (
    <Transition show={isOpen} appear>
      <Dialog as="div" className="relative z-50" onClose={() => !isLoading && onClose()}>
        <div className="fixed inset-0 bg-black/30 backdrop-blur-sm" aria-hidden="true" />
        <div className="fixed inset-0 flex items-center justify-center p-4">
          <Dialog.Panel className="mx-auto max-w-sm w-full bg-background rounded-2xl shadow-xl overflow-hidden border border-black/5 dark:border-white/5">
            <div className="flex items-center justify-between px-6 py-4 border-b border-black/5 dark:border-white/5">
              <Dialog.Title className="text-lg font-semibold text-foreground flex items-center gap-2">
                {isDestructive && <Alert01Icon className="w-5 h-5 text-red-500" />}
                {title}
              </Dialog.Title>
              {!isLoading && (
                <button onClick={onClose} className="text-muted-foreground hover:text-foreground transition-colors">
                  <Cancel01Icon className="w-5 h-5" />
                </button>
              )}
            </div>

            <div className="p-6">
              <p className="text-sm text-muted-foreground mb-6">
                {description}
              </p>
              
              <div className="flex justify-end gap-3">
                <button
                  onClick={onClose}
                  disabled={isLoading}
                  className="px-4 py-2 text-sm font-medium text-foreground hover:bg-black/5 dark:hover:bg-white/5 rounded-full transition-colors disabled:opacity-50"
                >
                  Cancel
                </button>
                <button
                  onClick={onConfirm}
                  disabled={isLoading}
                  className={`px-4 py-2 text-sm font-medium rounded-full transition-colors disabled:opacity-50 ${
                    isDestructive 
                      ? "bg-red-500 hover:bg-red-600 text-white" 
                      : "bg-primary hover:bg-primary/90 text-primary-foreground"
                  }`}
                >
                  {isLoading ? "Processing..." : confirmText}
                </button>
              </div>
            </div>
          </Dialog.Panel>
        </div>
      </Dialog>
    </Transition>
  );
}
