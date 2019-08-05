require 'json'

providers = JSON.parse(File.read('providers.json'))

working_providers = providers.map do |p|
  url = p["url"]
  if system("curl --connect-timeout 2 -so/dev/null --fail #{url}")
    p
  else next
  end
end

puts JSON.dump(working_providers.compact)
