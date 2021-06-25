// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';
import * as child_process from 'child_process';
import { pullFromK8s } from './pull';
import { autoImport } from './autoImport';
import { lint } from './lint';

export function getExecPath(context: vscode.ExtensionContext) {
	var os: string = process.platform;
	if (os === "win32") {
		os = "windows";
	}
	var path = context.asAbsolutePath(`bin/ytt-lint-${os}`);
	if (os != "windows") {
		child_process.exec(`chmod +x "${path}"`);
	}
	return path;
}

export function getSchemaPath(context: vscode.ExtensionContext) {
	return context.asAbsolutePath(`schema`);
}

// this method is called when your extension is activated
// your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {

	// Use the console to output diagnostic information (console.log) and errors (console.error)
	// This line of code will only be executed once when your extension is activated
	console.log('Congratulations, your extension "ytt-lint" is now active!');
	let timeout: NodeJS.Timer | undefined = undefined;

	// The command has been defined in the package.json file
	// Now provide the implementation of the command with registerCommand
	// The commandId parameter must match the command field in package.json
	let disposablePull = vscode.commands.registerCommand('extension.pullSchemaK8S', _ => pullFromK8s(context));

	let disposableImport = vscode.commands.registerCommand('extension.importSchema', e => autoImport(context, e));

	let diagnosticCollection = vscode.languages.createDiagnosticCollection('ytt-lint');

	context.subscriptions.push(disposablePull);
	context.subscriptions.push(disposableImport);
	context.subscriptions.push(diagnosticCollection);

	function triggerUpdateDecorations() {
		if (timeout) {
			clearTimeout(timeout);
			timeout = undefined;
		}
		timeout = setTimeout(_ => {
			if (activeEditor != null) {
				lint(context, activeEditor, diagnosticCollection);
			}
		}, 500);
	}

	let activeEditor = vscode.window.activeTextEditor;

	if (activeEditor) {
		triggerUpdateDecorations();
	}

	vscode.window.onDidChangeActiveTextEditor(editor => {
		activeEditor = editor;
		if (editor) {
			triggerUpdateDecorations();
		}
	}, null, context.subscriptions);

	vscode.workspace.onDidChangeTextDocument(event => {
		if (activeEditor && event.document === activeEditor.document) {
			triggerUpdateDecorations();
		}
	}, null, context.subscriptions);
}

// this method is called when your extension is deactivated
export function deactivate() { }
