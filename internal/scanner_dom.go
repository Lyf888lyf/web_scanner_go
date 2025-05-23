package internal

// Page 或 Frame 都可以传入
func ExtractDOMResources(evalTarget interface {
	Evaluate(expression string, arg ...interface{}) (interface{}, error)
}) ([]string, error) {
	script := `
	() => {
		const urls = [];
		const push = (attr) => el => el[attr] && urls.push(el[attr]);

		document.querySelectorAll('script[src]').forEach(push('src'));
		document.querySelectorAll('link[rel=stylesheet]').forEach(push('href'));
		document.querySelectorAll('img[src]').forEach(push('src'));
		document.querySelectorAll('a[href]').forEach(push('href'));

		return urls;
	}`
	result, err := evalTarget.Evaluate(script)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		var list []string
		for _, val := range arr {
			if str, ok := val.(string); ok {
				list = append(list, str)
			}
		}
		return list, nil
	}
	return nil, nil
}
