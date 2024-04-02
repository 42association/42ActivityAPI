# Go Web Application

このプロジェクトは、Gin Webフレームワークを使用したGo言語によるWebアプリケーションです。認証フローを実装しており、ユーザーが42 APIを通じて自身のデータを取得できるようにします。

## 始め方

このプロジェクトを使用するには、以下の手順に従ってください。

### 必要条件

- Go言語がインストールされていること
- `git` がインストールされていること
- 42 APIへのアクセスと、クライアントID (`UID`) とクライアントシークレット (`SECRET`) を取得していること

### インストール方法

1. リポジトリをクローンします。

   ```bash
   git clone https://github.com/your-username/your-project-name.git
   cd your-project-name
   ```

2. `.env` ファイルをプロジェクトのルートに作成し、以下の環境変数を設定します。

   ```
   UID=your_42_api_client_id
   SECRET=your_42_api_client_secret
   CALLBACK_URL=your_callback_url
   ```

3. 依存関係をインストールします。

   ```bash
   go mod tidy
   ```

4. アプリケーションを実行します。

   ```bash
   go run .
   ```

   これにより、デフォルトで `localhost:8080` にWebサーバーが立ち上がります。

## 使用方法

アプリケーションが実行されているときに、ブラウザを開き `http://localhost:8080` にアクセスしてください。ユーザー認証を行い、認証が成功すると、ユーザーの42 Intra名が表示されます。

## 機能

- `.env` ファイルからの環境変数の読み込み
- 42 APIを使用したOAuth認証
- ユーザーデータの取得と表示
- エラーハンドリング

## 貢献

プルリクエストはいつでも歓迎です。大きな変更を考えている場合は、まずissueを開いて話し合ってください。

## ライセンス

[MIT](https://choosealicense.com/licenses/mit/)