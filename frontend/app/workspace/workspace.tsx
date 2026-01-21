// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import { AIPanel } from "@/app/aipanel/aipanel";
import logoSvg from "@/app/asset/logo.svg?url";
import { ErrorBoundary } from "@/app/element/errorboundary";
import { CenteredDiv } from "@/app/element/quickelems";
import { ModalsRenderer } from "@/app/modals/modalsrenderer";
import { TabBar } from "@/app/tab/tabbar";
import { TabContent } from "@/app/tab/tabcontent";
import { Widgets } from "@/app/workspace/widgets";
import { WorkspaceLayoutModel } from "@/app/workspace/workspace-layout-model";
import { atoms, createBlock, createTab, getApi } from "@/store/global";
import { fireAndForget } from "@/util/util";
import { useAtomValue } from "jotai";
import { memo, useEffect, useRef } from "react";
import {
    ImperativePanelGroupHandle,
    ImperativePanelHandle,
    Panel,
    PanelGroup,
    PanelResizeHandle,
} from "react-resizable-panels";

const WorkspaceElem = memo(() => {
    const workspaceLayoutModel = WorkspaceLayoutModel.getInstance();
    const tabId = useAtomValue(atoms.staticTabId);
    const ws = useAtomValue(atoms.workspace);
    const initialAiPanelPercentage = workspaceLayoutModel.getAIPanelPercentage(window.innerWidth);
    const panelGroupRef = useRef<ImperativePanelGroupHandle>(null);
    const aiPanelRef = useRef<ImperativePanelHandle>(null);
    const panelContainerRef = useRef<HTMLDivElement>(null);
    const aiPanelWrapperRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (aiPanelRef.current && panelGroupRef.current && panelContainerRef.current && aiPanelWrapperRef.current) {
            workspaceLayoutModel.registerRefs(
                aiPanelRef.current,
                panelGroupRef.current,
                panelContainerRef.current,
                aiPanelWrapperRef.current
            );
        }
    }, []);

    useEffect(() => {
        const isVisible = workspaceLayoutModel.getAIPanelVisible();
        getApi().setAinAiOpen(isVisible);
    }, []);

    useEffect(() => {
        window.addEventListener("resize", workspaceLayoutModel.handleWindowResize);
        return () => window.removeEventListener("resize", workspaceLayoutModel.handleWindowResize);
    }, []);

    return (
        <div className="flex flex-col w-full flex-grow overflow-hidden">
            <TabBar key={ws.oid} workspace={ws} />
            <div ref={panelContainerRef} className="flex flex-row flex-grow overflow-hidden">
                <ErrorBoundary key={tabId}>
                    <PanelGroup
                        direction="horizontal"
                        onLayout={workspaceLayoutModel.handlePanelLayout}
                        ref={panelGroupRef}
                    >
                        <Panel
                            ref={aiPanelRef}
                            collapsible
                            defaultSize={initialAiPanelPercentage}
                            order={1}
                            className="overflow-hidden"
                        >
                            <div ref={aiPanelWrapperRef} className="w-full h-full">
                                {tabId !== "" && <AIPanel />}
                            </div>
                        </Panel>
                        <PanelResizeHandle className="w-0.5 bg-transparent hover:bg-zinc-500/20 transition-colors" />
                        <Panel order={2} defaultSize={100 - initialAiPanelPercentage}>
                            {tabId === "" ? (
                                <CenteredDiv>
                                    <div className="flex flex-col items-center gap-6">
                                        <img
                                            src={logoSvg}
                                            alt="Ain Term Logo"
                                            style={{ width: "250px", height: "250px" }}
                                        />
                                        <div className="flex flex-wrap gap-3 justify-center max-w-md">
                                            <button
                                                onClick={() => {
                                                    createTab();
                                                    setTimeout(() => {
                                                        fireAndForget(() => createBlock({ meta: { view: "term" } }));
                                                    }, 100);
                                                }}
                                                className="flex items-center gap-2 px-4 py-2 bg-[#3BAFF7] hover:bg-[#2d8bc7] text-white rounded-md transition-colors"
                                            >
                                                <i className="fa-sharp fa-solid fa-terminal"></i>
                                                <span>Open Terminal</span>
                                            </button>
                                            <button
                                                onClick={() => {
                                                    createTab();
                                                    setTimeout(() => {
                                                        fireAndForget(() => createBlock({ meta: { view: "preview" } }));
                                                    }, 100);
                                                }}
                                                className="flex items-center gap-2 px-4 py-2 bg-[#3BAFF7] hover:bg-[#2d8bc7] text-white rounded-md transition-colors"
                                            >
                                                <i className="fa-sharp fa-solid fa-folder-open"></i>
                                                <span>File Explorer</span>
                                            </button>
                                            <button
                                                onClick={() => {
                                                    createTab();
                                                    setTimeout(() => {
                                                        fireAndForget(() => createBlock({ meta: { view: "sysinfo" } }));
                                                    }, 100);
                                                }}
                                                className="flex items-center gap-2 px-4 py-2 bg-[#3BAFF7] hover:bg-[#2d8bc7] text-white rounded-md transition-colors"
                                            >
                                                <i className="fa-sharp fa-solid fa-chart-line"></i>
                                                <span>System Info</span>
                                            </button>
                                            <button
                                                onClick={() => {
                                                    createTab();
                                                    setTimeout(() => {
                                                        const layoutModel = WorkspaceLayoutModel.getInstance();
                                                        layoutModel.setAIPanelVisible(true);
                                                    }, 100);
                                                }}
                                                className="flex items-center gap-2 px-4 py-2 bg-[#3BAFF7] hover:bg-[#2d8bc7] text-white rounded-md transition-colors"
                                            >
                                                <i className="fa-sharp fa-solid fa-sparkles"></i>
                                                <span>Open AI Panel</span>
                                            </button>
                                        </div>
                                    </div>
                                </CenteredDiv>
                            ) : (
                                <div className="flex flex-row h-full">
                                    <TabContent key={tabId} tabId={tabId} />
                                    <Widgets />
                                </div>
                            )}
                        </Panel>
                    </PanelGroup>
                    <ModalsRenderer />
                </ErrorBoundary>
            </div>
        </div>
    );
});

WorkspaceElem.displayName = "WorkspaceElem";

export { WorkspaceElem as Workspace };
