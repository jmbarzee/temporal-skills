import * as vscode from "vscode";
import * as path from "path";
import { execFile } from "child_process";
import { promisify } from "util";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  Executable,
} from "vscode-languageclient/node";

const execFileAsync = promisify(execFile);

let client: LanguageClient | undefined;

export function activate(context: vscode.ExtensionContext) {
  // Start LSP client
  startLanguageClient(context);

  // Register visualize command
  const visualizeCommand = vscode.commands.registerCommand(
    "twf.visualize",
    () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor || editor.document.languageId !== "twf") {
        vscode.window.showWarningMessage("Please open a .twf file to visualize");
        return;
      }
      WorkflowVisualizerPanel.createOrShow(context.extensionUri, editor.document);
    }
  );

  context.subscriptions.push(visualizeCommand);

  // Watch for document changes to update visualization
  context.subscriptions.push(
    vscode.workspace.onDidSaveTextDocument((doc) => {
      if (doc.languageId === "twf") {
        WorkflowVisualizerPanel.updateIfVisible(doc);
      }
    })
  );
}

function startLanguageClient(context: vscode.ExtensionContext) {
  const config = vscode.workspace.getConfiguration("twf.lsp");
  const configPath = config.get<string>("path", "");
  const command = configPath || "twf-lsp";

  const serverOptions: ServerOptions = {
    run: { command } as Executable,
    debug: { command } as Executable,
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: "file", language: "twf" }],
    outputChannelName: "TWF Language Server",
  };

  client = new LanguageClient(
    "twf-lsp",
    "TWF Language Server",
    serverOptions,
    clientOptions
  );

  client.start().catch((err) => {
    vscode.window.showWarningMessage(
      `Failed to start TWF language server: ${err.message}. ` +
        `Install it with: go install github.com/jmbarzee/temporal-skills/lsp/cmd/twf-lsp@latest`
    );
  });

  context.subscriptions.push({
    dispose: () => {
      if (client) {
        client.stop();
      }
    },
  });
}

export function deactivate(): Thenable<void> | undefined {
  if (client) {
    return client.stop();
  }
  return undefined;
}

/**
 * Manages workflow visualizer webview panels
 */
class WorkflowVisualizerPanel {
  public static currentPanel: WorkflowVisualizerPanel | undefined;
  public static readonly viewType = "twfVisualizer";

  private readonly _panel: vscode.WebviewPanel;
  private readonly _extensionUri: vscode.Uri;
  private _document: vscode.TextDocument;
  private _disposables: vscode.Disposable[] = [];

  public static createOrShow(
    extensionUri: vscode.Uri,
    document: vscode.TextDocument
  ) {
    const column = vscode.ViewColumn.Beside;

    // If we already have a panel, show it
    if (WorkflowVisualizerPanel.currentPanel) {
      WorkflowVisualizerPanel.currentPanel._panel.reveal(column);
      WorkflowVisualizerPanel.currentPanel._document = document;
      WorkflowVisualizerPanel.currentPanel._update();
      return;
    }

    // Create a new panel
    const panel = vscode.window.createWebviewPanel(
      WorkflowVisualizerPanel.viewType,
      "TWF Visualizer",
      column,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
        localResourceRoots: [
          vscode.Uri.joinPath(extensionUri, "dist", "webview"),
        ],
      }
    );

    WorkflowVisualizerPanel.currentPanel = new WorkflowVisualizerPanel(
      panel,
      extensionUri,
      document
    );
  }

  public static updateIfVisible(document: vscode.TextDocument) {
    if (WorkflowVisualizerPanel.currentPanel) {
      WorkflowVisualizerPanel.currentPanel._document = document;
      WorkflowVisualizerPanel.currentPanel._update();
    }
  }

  private constructor(
    panel: vscode.WebviewPanel,
    extensionUri: vscode.Uri,
    document: vscode.TextDocument
  ) {
    this._panel = panel;
    this._extensionUri = extensionUri;
    this._document = document;

    // Set initial HTML content
    this._panel.webview.html = this._getHtmlForWebview();

    // Listen for when the panel is disposed
    this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

    // Handle messages from the webview
    this._panel.webview.onDidReceiveMessage(
      (message) => {
        switch (message.type) {
          case "ready":
            this._update();
            break;
        }
      },
      null,
      this._disposables
    );
  }

  public dispose() {
    WorkflowVisualizerPanel.currentPanel = undefined;

    this._panel.dispose();

    while (this._disposables.length) {
      const x = this._disposables.pop();
      if (x) {
        x.dispose();
      }
    }
  }

  private async _update() {
    try {
      const ast = await this._parseDocument();
      this._panel.webview.postMessage({ type: "ast", data: ast });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      this._panel.webview.postMessage({ type: "error", message: errorMessage });
    }
  }

  private async _parseDocument(): Promise<unknown> {
    const config = vscode.workspace.getConfiguration("twf.parser");
    const configPath = config.get<string>("path", "");
    const parserCommand = configPath || "parse";

    // Get all .twf files in the workspace that might be needed for resolution
    const workspaceFolder = vscode.workspace.getWorkspaceFolder(this._document.uri);
    let twfFiles: string[] = [];
    
    if (workspaceFolder) {
      const pattern = new vscode.RelativePattern(workspaceFolder, "**/*.twf");
      const uris = await vscode.workspace.findFiles(pattern);
      twfFiles = uris.map((uri) => uri.fsPath);
    }
    
    // Ensure current file is included
    const currentPath = this._document.uri.fsPath;
    if (!twfFiles.includes(currentPath)) {
      twfFiles.push(currentPath);
    }

    try {
      const { stdout, stderr } = await execFileAsync(parserCommand, [
        "--json",
        ...twfFiles,
      ]);

      if (stderr) {
        console.warn("Parser stderr:", stderr);
      }

      return JSON.parse(stdout);
    } catch (err) {
      if (err instanceof Error && "stderr" in err) {
        throw new Error((err as { stderr: string }).stderr || err.message);
      }
      throw err;
    }
  }

  private _getHtmlForWebview(): string {
    const webview = this._panel.webview;

    // Get URIs for webview resources
    const scriptUri = webview.asWebviewUri(
      vscode.Uri.joinPath(this._extensionUri, "dist", "webview", "visualizer.js")
    );
    const styleUri = webview.asWebviewUri(
      vscode.Uri.joinPath(this._extensionUri, "dist", "webview", "visualizer.css")
    );

    // Use a nonce to only allow specific scripts
    const nonce = getNonce();

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource} 'unsafe-inline'; script-src 'nonce-${nonce}';">
    <link href="${styleUri}" rel="stylesheet">
    <title>TWF Workflow Visualizer</title>
    <style>
      html, body, #root {
        height: 100%;
        width: 100%;
        margin: 0;
        padding: 0;
        overflow: hidden;
      }
    </style>
</head>
<body class="vscode-dark">
    <div id="root"></div>
    <script nonce="${nonce}" src="${scriptUri}"></script>
</body>
</html>`;
  }
}

function getNonce(): string {
  let text = "";
  const possible =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  for (let i = 0; i < 32; i++) {
    text += possible.charAt(Math.floor(Math.random() * possible.length));
  }
  return text;
}
