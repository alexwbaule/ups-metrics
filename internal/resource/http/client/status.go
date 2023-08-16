package client

var ErrorCodes = func(code string) string {
	status := map[string]string{
		"S001": "Sucesso",
		"S002": "Usuario ou senha inválidos",
		"S003": "Access Token inválido",
		"S004": "",
		"S005": "Comando inválido",
		"S006": "",
		"S007": "",
		"S016": "Usuário não está autenticado",
		"S017": "Usuário não tem permissão",
		"S018": "Numero máximo de usuarios criados",
		"S019": "Este usuário já existe",
		"S020": "Usuário não encontrado",
		"S021": "Não é possível excluir todos os administradores",
		"S032": "Não foi possível gravar a imagem",
		"S033": "A imagem utilizada não é válida",
		"S034": "Erro ao utilizar imagem gravada",
		"S035": "Não é possível voltar para o firmware de fábrica",
		"S036": "Não é possível apagar a memória flash",
		"S048": "Não foi possível desencriptar dados",
		"S255": "Erro desconhecido",
	}
	return status[code]
}
