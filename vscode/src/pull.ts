import * as vscode from 'vscode';
import * as child_process from 'child_process';
import * as YAML from 'yaml';
import { promises as fs } from 'fs';
import { getExecPath } from './extension';

export async function pullFromK8s(extensionContext: vscode.ExtensionContext) {
    const EXEC_PATH = getExecPath(extensionContext);

    let kubeconfigEnv = process.env["KUBECONFIG"] ?? "~/.kube/config";

    const USE_CURRENT = `use current kubeconfig (${kubeconfigEnv})`;
    const SELECT = "select kubeconfig from filesystem";
    let mode = await vscode.window.showQuickPick([USE_CURRENT, SELECT]);

    if (mode == undefined) {
        return;
    }

    let kubeconfigPath: string;
    if (mode == SELECT) {
        var kubeconfig = await vscode.window.showOpenDialog({
            filters: {
                "all": ["*"],
                "kubeconfig": ["kubeconfig"],
                "yaml": ["yaml"],
            },
            defaultUri: vscode.Uri.file(kubeconfigEnv),
            openLabel: "Use",
        });
        if (kubeconfig == undefined) {
            return;
        }
        kubeconfigPath = kubeconfig[0].path;
    } else {
        kubeconfigPath = kubeconfigEnv;
    }


    interface contextSelectionItem extends vscode.QuickPickItem {
        context: string;
    }

    let content = await fs.readFile(kubeconfigPath, { encoding: "utf-8" });
    let parsedKubeconfig = YAML.parse(content);
    let context: string;
    if (parsedKubeconfig.contexts.length == 1) {
        context = parsedKubeconfig.contexts[0].name;
    } else {
        let contexts: contextSelectionItem[] = parsedKubeconfig.contexts.map((x: any) => {
            var suffix = "";
            if (parsedKubeconfig["current-context"] == x.name) {
                suffix = " (current)";
            }
            return {
                label: `Context: ${x.name} ${suffix}`,
                context: x.name,
            };
        });

        let selectedContext = await vscode.window.showQuickPick(contexts);
        if (selectedContext == undefined) {
            return;
        }
        context = selectedContext.context;
    }

    let out = vscode.window.createOutputChannel(`ytt-lint schema pull`);
    let exec = child_process.execFile(EXEC_PATH, ['--pull-from-k8s', '--kubeconfig', kubeconfigPath, '--context', context]);

    exec.stdout?.on('data', (data) => {
        out.append(data);
    });

    exec.stderr?.on('data', (data) => {
        out.append(data);
    });

    exec.on('close', (code) => {
        out.appendLine(`Exit code: ${code}`);
    });

    out.show();
}
