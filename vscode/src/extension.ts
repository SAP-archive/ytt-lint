// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';
import * as child_process from 'child_process';
import * as YAML from 'yaml';
import * as path from 'path';
import { promises as fs } from 'fs';

function getExecPath(context: vscode.ExtensionContext) {
	var os: string = process.platform;
	if (os  === "win32") {
		os = "windows";
	}
	var path = context.asAbsolutePath(`bin/ytt-lint-${os}`);
	if (os != "windows") {
		child_process.exec(`chmod +x "${path}"`);
	}
	return path;
}

// this method is called when your extension is activated
// your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {

	// Use the console to output diagnostic information (console.log) and errors (console.error)
	// This line of code will only be executed once when your extension is activated
	console.log('Congratulations, your extension "ytt-lint" is now active!');
	let timeout: NodeJS.Timer | undefined = undefined;

	const SCHEMA_PATH = context.asAbsolutePath(`schema`);
	const EXEC_PATH = getExecPath(context);

	// The command has been defined in the package.json file
	// Now provide the implementation of the command with registerCommand
	// The commandId parameter must match the command field in package.json
	let disposablePull = vscode.commands.registerCommand('extension.pullSchemaK8S', async () => {

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

		let content = await fs.readFile(kubeconfigPath, {encoding: "utf-8"});
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
	});

	let disposableImport = vscode.commands.registerCommand('extension.importSchema', async (e) => {
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
			env: Object.assign({YTT_LINT_SCHEMA_PATH: SCHEMA_PATH}, process.env)
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
	});

	let diagnosticCollection = vscode.languages.createDiagnosticCollection('ytt-lint');
	
	context.subscriptions.push(disposablePull);
	context.subscriptions.push(disposableImport);
	context.subscriptions.push(diagnosticCollection);

	function lint() {
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
			env: Object.assign({YTT_LINT_SCHEMA_PATH: SCHEMA_PATH}, process.env)
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

	function triggerUpdateDecorations() {
		if (timeout) {
			clearTimeout(timeout);
			timeout = undefined;
		}
		timeout = setTimeout(lint, 500);
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
export function deactivate() {}
