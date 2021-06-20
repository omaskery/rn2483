package fake

func invalidParam(ctx *commandContext) error {
	return ctx.writeResponse("invalid_param")
}

func ok(ctx *commandContext) error {
	return ctx.writeResponse("ok")
}
