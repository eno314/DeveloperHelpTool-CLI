# Five Lines of Code Policy Integration

## Problem Statement
どのようにすれば、「Five Lines of Code」の主要な原則（関数の短小化、ネストの排除、elseの撲滅など）を、Go言語の特性（エラーハンドリングなど）やプロジェクトの既存方針と矛盾しない形でポリシーに統合し、コードの保守性を最大化できるか？

## Recommended Direction
ポリシーに「8. Function Size and Control Flow」を新設し、関数サイズ制限（実質5行、全体10行以内）、`else` の禁止、早期リターン（Guard Clauses）の徹底、ネストレベルの制限（最大1階層）を導入します。
これにより、開発者は画面スクロールなしに関数の全体像を把握できるようになり、コードの可読性と保守性が向上します。

## Key Assumptions to Validate
- [ ] **可読性の維持**: 関数を細分化することで、かえって処理の全体像が追いにくくならないか（既存コードへの適用を通じて検証）。
- [ ] **エラーハンドリング除外の運用**: レビュー時（`code-review-and-quality`）に「何を行数カウントから除外するか」の認識合わせがスムーズに行われるか。

## MVP Scope
- `.agents/implementation_policy.md` にルールを追加し、今後の開発およびリファクタリングで適用する。
- 既存のロジック（例：探索アルゴリズムやCLIコマンド）の一部にこのルールを適用し、効果を実証する。

## Not Doing (and Why)
- **Goのイディオムに反するポリモーフィズムの強制**:
  - 『Five Lines of Code』本で推奨されている「switchやifのポリモーフィズムへの置き換え」は、Goではコードが過度に複雑化するため本ポリシーからは除外する（Goにおいては `switch` や `map` によるディスパッチの方がシンプルで好ましいため）。

## Open Questions
- 特になし（ユーザー合意済み）。
