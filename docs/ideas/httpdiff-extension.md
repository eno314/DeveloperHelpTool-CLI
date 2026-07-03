# httpdiff 機能拡張 構想案

## 1. 課題ステートメント (Problem Statement)
リファクタリング後やシステム移行時において、POSTリクエストやカスタムヘッダー、JSON以外のレスポンス（XML/HTML）を伴うAPIの新旧差異を、CLIから手軽かつ詳細に比較・検証できるようにする。

## 2. 推奨される方向性 (Recommended Direction)
既存の `httpdiff` コマンドラインツールをベースに、以下の新規オプションを追加して機能を直接拡張します。

- `-method <METHOD>`: HTTPメソッドの指定（デフォルト: `GET`）。`POST` 等に対応。
- `-body <STRING>`: POST等のリクエストボディを直接文字列で指定。
- `-body-file <PATH>`: リクエストボディを記述した外部ファイルを読み込んで送信（`-body` とは排他）。
- `-header <NAME: VALUE>`: 任意のHTTPヘッダーを複数指定可能にする（例: `-header "Content-Type: application/json" -header "Authorization: Bearer token"`）。

レスポンス比較は、JSON/XML/HTMLに関わらず、一旦は単純なプレーンテキストとしての差分比較（`cmp.Diff`）を行います。

## 3. 検証すべき前提条件 (Key Assumptions to Validate)
- [ ] Goの標準 `flag` パッケージでカスタム型（`flag.Value`）を実装し、同一の `-header` フラグを複数回指定してパースできること。
- [ ] XMLやHTMLのインデントや改行コードの違いもプレーンテキストとしてそのまま差分出力されるため、それらが許容可能なノイズレベルであること。
- [ ] `-body` と `-body-file` が同時に指定された場合に、適切にエラー（排他制御）にできること。

## 4. MVPスコープ (MVP Scope)

### ✅ 対応すること (IN)
- `GET`/`POST` などの任意のHTTPメソッドでのリクエスト送信。
- 複数ヘッダーの指定（指定されたヘッダーは `host1` と `host2` の両方に共通で送信）。
- インライン文字列によるリクエストボディ指定 (`-body`)。
- 外部ファイルからリクエストボディをロードする機能 (`-body-file`)。
- レスポンスボディのプレーンテキストでの差分比較。

### ❌ 対応しないこと (OUT)
- `host1` と `host2` で異なるヘッダーを送信する機能（トークンの出し分け等）。
- リクエストパスごとに個別のボディを指定する機能（全パス共通ボディのみサポート）。
- XML/HTMLやJSONの自動整形（Prettify）によるセマンティックな比較（インデント差異の無視など）。
- 設定ファイル（YAML/JSON）によるバッチ実行機能。
