// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import settingsSchema from "../../../schema/settings.json";
import connectionsSchema from "../../../schema/connections.json";
import aipresetsSchema from "../../../schema/aipresets.json";
import bgpresetsSchema from "../../../schema/bgpresets.json";
import ainaiSchema from "../../../schema/ainai.json";
import widgetsSchema from "../../../schema/widgets.json";

type SchemaInfo = {
    uri: string;
    fileMatch: Array<string>;
    schema: object;
};

const MonacoSchemas: SchemaInfo[] = [
    {
        uri: "wave://schema/settings.json",
        fileMatch: ["*/ainconfigPATH/settings.json"],
        schema: settingsSchema,
    },
    {
        uri: "wave://schema/connections.json",
        fileMatch: ["*/ainconfigPATH/connections.json"],
        schema: connectionsSchema,
    },
    {
        uri: "wave://schema/aipresets.json",
        fileMatch: ["*/ainconfigPATH/presets/ai.json"],
        schema: aipresetsSchema,
    },
    {
        uri: "wave://schema/bgpresets.json",
        fileMatch: ["*/ainconfigPATH/presets/bg.json"],
        schema: bgpresetsSchema,
    },
    {
        uri: "wave://schema/ainai.json",
        fileMatch: ["*/ainconfigPATH/ainai.json"],
        schema: ainaiSchema,
    },
    {
        uri: "wave://schema/widgets.json",
        fileMatch: ["*/ainconfigPATH/widgets.json"],
        schema: widgetsSchema,
    },
];

export { MonacoSchemas };
