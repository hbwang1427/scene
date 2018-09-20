rm group*
rm model.json
tensorflowjs_converter --input_format keras ./weights_porcelain.h5 ./
