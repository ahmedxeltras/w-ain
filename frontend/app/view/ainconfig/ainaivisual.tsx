// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import type { AinConfigViewModel } from "@/app/view/ainconfig/ainconfig-model";
import { memo } from "react";

interface AinAiVisualContentProps {
    model: AinConfigViewModel;
}

export const AinAiVisualContent = memo(({ model }: AinAiVisualContentProps) => {
    return (
        <div className="flex flex-col gap-4 p-6 h-full">
            <div className="text-lg font-semibold">Ain AI Modes - Visual Editor</div>
            <div className="text-muted-foreground">Visual editor coming soon...</div>
        </div>
    );
});

AinAiVisualContent.displayName = "AinAiVisualContent";
