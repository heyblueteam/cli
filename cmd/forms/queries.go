package forms

// Shared GraphQL fragments and queries used by multiple subcommands.

const formDetailFragment = `
	fragment FormDetailFields on Form {
		id
		uid
		title
		description
		footerText
		showFooter
		isActive
		theme
		primaryColor
		hideBranding
		responseText
		submitText
		imageURL
		redirectURL
		snapshotURL
		createdAt
		updatedAt
		todoList {
			id
			title
		}
		formFields(orderBy: position_ASC) {
			id
			uid
			field
			name
			placeholder
			required
			position
			addToDescription
			hidden
			extraInfo
			customField {
				id
				name
				type
			}
		}
	}
`
