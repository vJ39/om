OM
==

## これは何?

## parsekmz
- Google Map「マイマップ」エクスポートファイル(.kmz)からTSVパース
- XML -> json変換

### kmzダウンロード方法
- マイマップのダウンロード時、チェック2つとも入れずにダウンロード。
- マイマップ名.kmz ファイルがダウンロードできます。

### プログラム実行

```
% ./parsekmz -f master.kmz
```
#### オプション
- -f
    - tsvファイル名
    - デフォルト: master.tsv

#### 解説
- master.kmz.json が生成されます。
- 標準出力にtsv形式で以下情報を出力します。
    - Name(プロットした名称)
    - Latitude(軽度)
    - Longitude(緯度)
    - Description, - プロットの詳細テキスト
        - インラインの画像や動画も含まれます。
    - ExtendedData.Data.Value,
        - 添付ファイルのURLが含まれます。
        - Description で利用されます。
    - StyleUrl,
        - 定義されたアイコンのid属性が入ります。
        - 定義部分はパースしてないので、識別用の参考値程度です。
- 高度データはありません。
- tsv形式はスプレッドシート等にペーストして利用します。

## buildkmz
- TSVデータからkmzファイル生成

### プログラム実行

```
% ./buildkmz
% ./buildkmz -f foo.tsv
```

- master.kmz ファイルが生成されます。

#### オプション
- -f
    - tsvファイル名
    - デフォルト: master.tsv

