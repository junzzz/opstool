# マスタ変換君

## これは何ですか？
マスターデータをバッチで読み込ませるcsvに変換するやつ


## 使い方


**注意**
以下のデータがローカルに必要です

- attribute_big_teaching_units
- attribute_difficulties
- attribute_middle_teaching_units
- attribute_school_ages
- attribute_small_teaching_units
- attribute_subject_categories
- attribute_subjects

DBへの接続は

- DBname:classi_development
- user:root
- pass:なし
- host:localhost
- port:3306

に固定してあるので、変えたい場合はmain.goの21〜25あたり変えてね

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
$ cd video_master_converter
$ glide get
$ go build -o video_master_converter main.go

# コンパイルしないで直接実行でも可
$ go run main.go -master ./master.csv
```

実行
```
$ ./video_master_converter -master ./master.csv
```

同じディレクトリにconvert.csvが作成されるので、エラーが出てないか確認する

### 引数

```
  -master string
        マスターファイル
```


### おまけ
1900行のcsvの変換にかかる時間

```
$ /usr/bin/time ./video_master_converter -master ~/Downloads/video_master_sample1910.csv
        0.12 real         0.02 user         0.00 sys
```
