# 教材動画書換君

## これは何ですか？
教材動画のURLを書き換えるツール

## 使い方
ローカルに環境変数を追加する

```
export AWS_SECRET_ACCESS_KEY=xxx
export AWS_ACCESS_KEY_ID=xxx
```

go1.6いれる

```
$ brew install go
$ echo 'export GOPATH=$HOME' >> .bashrc
```

glideが必要なのでbrewで入れる

```
$ brew install glide
```


コンパイルする
```
$ cd classi_opstool/video_link_editor
$ glide get
$ go build -o video_link_editor main.go video_link.go

# コンパイルしないで直接実行でも可
$ go run main.go video_link.go
```

### 単体の場合
バイナリファイルを実行する

```
$ ./video_link_editor -entry_cd 003a8f7f48bf7f87a429cf32d0494f1c6a5cbaaa -url https://hogehoge.com -stage production -dry-run=false
```

### 複数の場合
一覧のCSVファイルを用意

```test.csv
003a8f7f48bf7f87a429cf32d0494f1c6a5cbaaa,https://hogehoge.com
003a8f7f48bf7f87a429cf32d0494f1c6a5cbaaa,https://hogehoge.com
```

ファイルを指定して実行
```
$ ./video_link_editor_darwin -list ./test.csv -stage production -dry-run=false
```

### 引数

```
  -concurrent
       concurrent
  -dry-run
       dry run (default true)
  -entry_cd string
       cbankEntry.entry_cd
  -list string
       entry_cdと変更後URLのリストのcsvファイル
  -stage string
       env[STAGE] (default "staging")
  -url string
       after url
```
