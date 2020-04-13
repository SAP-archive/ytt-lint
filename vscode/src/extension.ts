// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';
import * as child_process from 'child_process';

// this method is called when your extension is activated
// your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {

	// Use the console to output diagnostic information (console.log) and errors (console.error)
	// This line of code will only be executed once when your extension is activated
	console.log('Congratulations, your extension "ytt-lint" is now active!');
	let timeout: NodeJS.Timer | undefined = undefined;

	const SCHEMA_PATH = context.asAbsolutePath(`schema`);
	const EXEC_PATH = context.asAbsolutePath(`bin/ytt-lint`);
	
	if (process.platform != "win32") {
		child_process.exec(`chmod +x "${EXEC_PATH}"`)
	}

	// The command has been defined in the package.json file
	// Now provide the implementation of the command with registerCommand
	// The commandId parameter must match the command field in package.json
	let disposable = vscode.commands.registerCommand('extension.pullSchemaK8S', () => {
		let out = vscode.window.createOutputChannel(`ytt-lint schema pull`);
		let exec = child_process.execFile(EXEC_PATH, ['--pull-from-k8s']);

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
	
	context.subscriptions.push(disposable);
	context.subscriptions.push(diagnosticCollection);

	function lint() {
		if (activeEditor == null) {
			return;
		}
		if (activeEditor.document.languageId != "yaml") {
			return;
		}
		let doc = activeEditor.document;

		// TODO: don't use '-f -' if file is saved
		let yaml = doc.getText();

		console.log('Running lint now!');

		diagnosticCollection.clear();
		let diagnosticMap: Map<string, vscode.Diagnostic[]> = new Map();

		// TODO: use spwan and then stream
		let linter = child_process.execFile(EXEC_PATH, ['-f', `-:${doc.fileName}`, '-o', 'json'], {
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
