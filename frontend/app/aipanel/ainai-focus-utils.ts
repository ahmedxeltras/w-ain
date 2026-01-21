// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

export function findAinAiPanel(element: HTMLElement): HTMLElement | null {
    let current: HTMLElement = element;
    while (current) {
        if (current.hasAttribute("data-ainai-panel")) {
            return current;
        }
        current = current.parentElement;
    }
    return null;
}

export function ainAiHasFocusWithin(focusTarget?: Element | null): boolean {
    if (focusTarget !== undefined) {
        if (focusTarget instanceof HTMLElement) {
            return findAinAiPanel(focusTarget) != null;
        }
        return false;
    }

    const focused = document.activeElement;
    if (focused instanceof HTMLElement) {
        const ainaiPanel = findAinAiPanel(focused);
        if (ainaiPanel) return true;
    }

    const sel = document.getSelection();
    if (sel && sel.anchorNode && sel.rangeCount > 0 && !sel.isCollapsed) {
        let anchor = sel.anchorNode;
        if (anchor instanceof Text) {
            anchor = anchor.parentElement;
        }
        if (anchor instanceof HTMLElement) {
            const ainaiPanel = findAinAiPanel(anchor);
            if (ainaiPanel) return true;
        }
    }

    return false;
}

export function ainAiHasSelection(): boolean {
    const sel = document.getSelection();
    if (!sel || sel.rangeCount === 0 || sel.isCollapsed) {
        return false;
    }

    let anchor = sel.anchorNode;
    if (anchor instanceof Text) {
        anchor = anchor.parentElement;
    }
    if (anchor instanceof HTMLElement) {
        return findAinAiPanel(anchor) != null;
    }

    return false;
}
