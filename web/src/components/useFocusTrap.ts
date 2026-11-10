import { useEffect, useRef, type RefObject } from "react";

const focusableSelector =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

function focusables(container: HTMLElement): HTMLElement[] {
  return [...container.querySelectorAll<HTMLElement>(focusableSelector)];
}

function initialFocus(container: HTMLElement): HTMLElement {
  return container.querySelector<HTMLElement>("[data-autofocus]") ?? focusables(container)[1] ?? focusables(container)[0] ?? container;
}

function wrapFocus(event: KeyboardEvent, container: HTMLElement): void {
  const items = focusables(container);
  const first = items[0];
  const last = items[items.length - 1];
  if (!first || !last) {
    event.preventDefault();
    return;
  }
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault();
    last.focus();
    return;
  }
  if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault();
    first.focus();
  }
}

export function useFocusTrap(ref: RefObject<HTMLElement | null>, onEscape: () => void): void {
  const escape = useRef(onEscape);

  useEffect(() => {
    escape.current = onEscape;
  }, [onEscape]);

  useEffect(() => {
    const container = ref.current;
    if (!container) return;
    const previous = document.activeElement as HTMLElement | null;
    initialFocus(container).focus();
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.stopPropagation();
        escape.current();
        return;
      }
      if (event.key === "Tab") wrapFocus(event, container);
    };
    container.addEventListener("keydown", onKeyDown);
    return () => {
      container.removeEventListener("keydown", onKeyDown);
      previous?.focus?.();
    };
  }, [ref]);
}
