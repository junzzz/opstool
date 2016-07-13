# しゅぞうBot

## 使い方

goとかglideいれる

```
$ brew install go glide
```

### 通常版
1. 42行めのAPP_TOKENにfacebookのAPP_TOKENいれる

2. コンパイル

```
$ go build syuzou.go
```

### Lambda版
1. 42行めのAPP_TOKENにfacebookのAPP_TOKENいれる

2. 依存ライブラリインストール

```
$ glide get
```

3. コンパイルとアップロードするためにzipにする

```
$ GOOS=linux go build lambda_syuzou.go
$ zip -r lambda.zip index.js lambda_syuzou
```
