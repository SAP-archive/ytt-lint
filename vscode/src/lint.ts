import * as vscode from 'vscode';
import * as child_process from 'child_process';
import { getExecPath, getSchemaPath } from './extension';

export function lint(context: vscode.ExtensionContext, activeEditor: vscode.TextEditor, diagnosticCollection: vscode.DiagnosticCollection) {
    const EXEC_PATH = getExecPath(context);
    const SCHEMA_PATH = getSchemaPath(context);

    if (activeEditor == null) {
        return;
    }
    if (["yaml", "ytt"].indexOf(activeEditor.document.languageId) < 0) {
        return;
    }
    let doc = activeEditor.document;

    // TODO: don't use '-f -' if file is saved
    let yaml = doc.getText();
    let root = vscode.workspace.getWorkspaceFolder(doc.uri)?.uri.path;

    console.log('Running lint now!');

    diagnosticCollection.clear();
    let diagnosticMap: Map<string, vscode.Diagnostic[]> = new Map();

    let args = ['-f', `-:${doc.fileName}`, '-o', 'json'];
    if (root) {
        args.push('--root');
        args.push(root);
    }
    // TODO: use spwan and then stream
    let linter = child_process.execFile(EXEC_PATH, args, {
        env: Object.assign({ YTT_LINT_SCHEMA_PATH: SCHEMA_PATH }, process.env)
    }, (error, stdout, stderr) => {
        console.log('Done linting:', error, stdout, stderr);
        let errors = JSON.parse(stdout);

        errors.forEach((error: { pos: string; msg: string; }) => {
            let [file, l] = error.pos.split(":");
            if (file != doc.fileName) {
                return;
            }
            if (l == undefined) {
                vscode.window.showErrorMessage(`ytt-lint has a bug: "${error.msg}" has no line info. Please open an issue.`);
                return;
            }
            let lineNum = parseInt(l) - 1;
            //let canonicalFile = vscode.Uri.file(file).toString();
            let canonicalFile = doc.uri.toString();

            let line = doc.lineAt(lineNum);
            let start = line.firstNonWhitespaceCharacterIndex;
            let end = line.range.end.character;
            //let range = line.range;
            //range.start = line.firstNonWhitespaceCharacterIndex;

            let range = new vscode.Range(lineNum, start, lineNum, end);
            let diagnostics = diagnosticMap.get(canonicalFile);
            if (!diagnostics) { diagnostics = []; }
            let diag = new vscode.Diagnostic(range, error.msg /*TODO: , error.severity*/);
            diag.source = "ytt-lint";
            diagnostics.push(diag);
            diagnosticMap.set(canonicalFile, diagnostics);
        });
        diagnosticMap.forEach((diags, file) => {
            diagnosticCollection.set(vscode.Uri.parse(file), diags);
        });

    });
    linter.stdin?.write(yaml);
    linter.stdin?.end();
}
