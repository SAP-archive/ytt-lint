import * as vscode from 'vscode';
import * as child_process from 'child_process';
import * as path from 'path';
import { getExecPath, getSchemaPath } from './extension';

export async function autoImport(context: vscode.ExtensionContext, e: any) {
    const EXEC_PATH = getExecPath(context);
    const SCHEMA_PATH = getSchemaPath(context);

    let importPath: string
    if (e) {
        importPath = e.fsPath;
    } else {
        if (vscode.window.activeTextEditor === undefined) {
            vscode.window.showErrorMessage("ytt-lint can't import schema: no editor window open/active");
            return;
        }
        importPath = vscode.window.activeTextEditor?.document.uri.fsPath;
    }

    let out = vscode.window.createOutputChannel(`ytt-lint schema import`);
    let exec = child_process.execFile(EXEC_PATH, ['--autoimport', '-f', importPath, '--root', path.dirname(importPath)], {
        env: Object.assign({ YTT_LINT_SCHEMA_PATH: SCHEMA_PATH }, process.env)
    });

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
