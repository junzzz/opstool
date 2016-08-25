const MAX_FAILS = 4;
var child_process = require('child_process');

exports.handler = function(event, context) {
  var srcKey    =
    decodeURIComponent(event.Records[0].s3.object.key.replace(/\+/g, " "));

  var proc = child_process.spawn('./make_thumb', [ srcKey || '' ], { stdio: 'inherit' })
  proc.on('close', function(code) {
    if(code !== 0) {
      return context.done(new Error('Process exited with non-zero status code'))
    }

    context.done(null)
  })
}
